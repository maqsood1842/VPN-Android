package com.vpn.core.vpn

import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.net.ConnectivityManager
import android.net.Network
import android.net.NetworkCapabilities
import android.net.NetworkRequest
import android.net.VpnService
import android.os.Build
import android.os.ParcelFileDescriptor
import android.util.Log
import com.wireguard.android.backend.GoBackend
import com.wireguard.config.Config
import com.wireguard.config.InterfaceConfig
import com.wireguard.config.PeerConfig
import com.wireguard.config.PeerConfig.Builder as PeerConfigBuilder
import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class VpnConnectionService @Inject constructor(
    private val context: Context
) {
    private var vpnInterface: ParcelFileDescriptor? = null
    private var goBackend: GoBackend? = null
    private val scope = CoroutineScope(Dispatchers.IO)
    private var isConnected = false
    private var killSwitchEnabled = true
    private var excludedApps: Set<String> = emptySet()
    private var includedApps: Set<String> = emptySet()
    private var connectivityManager: ConnectivityManager? = null
    private var networkCallback: ConnectivityManager.NetworkCallback? = null

    interface ConnectionCallback {
        fun onConnected()
        fun onDisconnected()
        fun onError(error: String)
    }

    private var connectionCallback: ConnectionCallback? = null

    fun setConnectionCallback(callback: ConnectionCallback?) {
        connectionCallback = callback
    }

    fun setKillSwitchEnabled(enabled: Boolean) {
        killSwitchEnabled = enabled
        if (enabled && isConnected) {
            enableKillSwitch()
        } else {
            disableKillSwitch()
        }
    }

    fun setSplitTunneling(excludedPackages: Set<String>, includedPackages: Set<String>) {
        excludedApps = excludedPackages
        includedApps = includedPackages
        if (isConnected) {
            applySplitTunneling()
        }
    }

    suspend fun connect(peerConfig: com.vpn.core.data.Models.PeerConfig): Result<Unit> {
        val deferred = CompletableDeferred<Result<Unit>>()
        scope.launch {
            try {
                if (isConnected) {
                    disconnect()
                }

                goBackend = GoBackend(context)

                val builder = VpnService.Builder()
                    .setSession("VPN Tunnel")
                    .setMtu(1420)
                    .addAddress("10.0.0.2", 32)
                    .addAddress("fd00::2", 128)
                    .addDnsServer("1.1.1.1")
                    .addDnsServer("1.0.0.1")
                    .addRoute("0.0.0.0", 0)
                    .addRoute("::", 0)

                if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.N) {
                    builder.setBlocking(true)
                }

                applySplitTunnelingToBuilder(builder)

                val intent = Intent(context, VpnService::class.java)
                val pendingIntent = PendingIntent.getActivity(
                    context, 0, intent, PendingIntent.FLAG_IMMUTABLE
                )
                builder.setConfigureIntent(pendingIntent)

                vpnInterface = builder.establish()

                if (vpnInterface == null) {
                    deferred.complete(Result.failure(Exception("Failed to establish VPN interface")))
                    return@launch
                }

                if (killSwitchEnabled) {
                    enableKillSwitch()
                }

                val tunnelConfig = buildTunnelConfig(peerConfig)
                
                goBackend?.createTunnel(tunnelConfig, object : GoBackend.Callback {
                    override fun onTunnelCreated() {
                        isConnected = true
                        connectionCallback?.onConnected()
                        deferred.complete(Result.success(Unit))
                    }

                    override fun onError(error: String) {
                        Log.e("VpnConnectionService", "Tunnel creation failed: $error")
                        cleanup()
                        connectionCallback?.onError(error)
                        deferred.complete(Result.failure(Exception(error)))
                    }
                })
            } catch (e: Exception) {
                Log.e("VpnConnectionService", "Connect failed", e)
                cleanup()
                connectionCallback?.onError(e.message ?: "Unknown error")
                deferred.complete(Result.failure(e))
            }
        }
        return deferred.await()
    }

    fun disconnect() {
        scope.launch {
            cleanup()
            connectionCallback?.onDisconnected()
        }
    }

    private fun cleanup() {
        try {
            disableKillSwitch()
            goBackend?.close()
            vpnInterface?.close()
            vpnInterface = null
            goBackend = null
            isConnected = false
        } catch (e: Exception) {
            Log.e("VpnConnectionService", "Disconnect failed", e)
        }
    }

    fun isConnected(): Boolean = isConnected

    fun isKillSwitchEnabled(): Boolean = killSwitchEnabled

    private fun applySplitTunnelingToBuilder(builder: VpnService.Builder) {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.LOLLIPOP) {
            if (includedApps.isNotEmpty()) {
                val pm = context.packageManager
                includedApps.forEach { packageName ->
                    try {
                        val uid = pm.getPackageUid(packageName, 0)
                        builder.addAllowedApplication(uid)
                    } catch (e: Exception) {
                        Log.w("VpnConnectionService", "Failed to add allowed app $packageName: $e")
                    }
                }
            } else if (excludedApps.isNotEmpty()) {
                val pm = context.packageManager
                excludedApps.forEach { packageName ->
                    try {
                        val uid = pm.getPackageUid(packageName, 0)
                        builder.addDisallowedApplication(uid)
                    } catch (e: Exception) {
                        Log.w("VpnConnectionService", "Failed to add disallowed app $packageName: $e")
                    }
                }
            }
        }
    }

    private fun applySplitTunneling() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.LOLLIPOP && vpnInterface != null) {
            // Note: On Android, split tunneling rules are set at establish time
            // For dynamic changes, we'd need to re-establish the VPN
            Log.i("VpnConnectionService", "Split tunneling rules applied at connection time")
        }
    }

    private fun enableKillSwitch() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.N && vpnInterface != null) {
            // The VPN interface is already blocking non-VPN traffic when setBlocking(true)
            // Additional kill switch: monitor connectivity and block if VPN drops
            startNetworkMonitoring()
        }
    }

    private fun disableKillSwitch() {
        stopNetworkMonitoring()
    }

    private fun startNetworkMonitoring() {
        connectivityManager = context.getSystemService(Context.CONNECTIVITY_SERVICE) as ConnectivityManager
        
        networkCallback = object : ConnectivityManager.NetworkCallback() {
            override fun onLost(network: Network) {
                if (isConnected && killSwitchEnabled) {
                    // Check if the lost network is our VPN
                    val caps = connectivityManager?.getNetworkCapabilities(network)
                    if (caps?.hasTransport(NetworkCapabilities.TRANSPORT_VPN) == true) {
                        Log.w("VpnConnectionService", "VPN network lost - kill switch activating")
                        scope.launch {
                            cleanup()
                            connectionCallback?.onError("VPN connection lost - kill switch activated")
                        }
                    }
                }
            }
        }

        val request = NetworkRequest.Builder()
            .addTransportType(NetworkCapabilities.TRANSPORT_VPN)
            .build()
        
        connectivityManager?.registerNetworkCallback(request, networkCallback!!)
    }

    private fun stopNetworkMonitoring() {
        connectivityManager?.let { cm ->
            networkCallback?.let { cb ->
                try {
                    cm.unregisterNetworkCallback(cb)
                } catch (e: Exception) {
                    Log.w("VpnConnectionService", "Error unregistering network callback: $e")
                }
            }
            networkCallback = null
        }
    }

    private fun buildTunnelConfig(peerConfig: com.vpn.core.data.Models.PeerConfig): String {
        val interfaceConfig = InterfaceConfig.Builder()
            .setPrivateKey(peerConfig.interface.privateKey)
            .addAddress("10.0.0.2/32")
            .addAddress("fd00::2/128")
            .setDns(peerConfig.interface.dns.joinToString(","))
            .build()

        val peer = PeerConfigBuilder()
            .setPublicKey(peerConfig.peer.publicKey)
            .setPresharedKey(peerConfig.peer.presharedKey)
            .addAllowedIp(peerConfig.peer.allowedIPs.joinToString(","))
            .setEndpoint(peerConfig.peer.endpoint)
            .setPersistentKeepalive(peerConfig.peer.persistentKeepalive)
            .build()

        val config = Config.Builder()
            .setInterface(interfaceConfig)
            .addPeer(peer)
            .build()

        return config.toString()
    }
}