package com.vpn.core.vpn

import android.app.PendingIntent
import android.content.Context
import android.content.Intent
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

                val tunnelConfig = buildTunnelConfig(peerConfig)
                
                goBackend?.createTunnel(tunnelConfig, object : GoBackend.Callback {
                    override fun onTunnelCreated() {
                        isConnected = true
                        deferred.complete(Result.success(Unit))
                    }

                    override fun onError(error: String) {
                        Log.e("VpnConnectionService", "Tunnel creation failed: $error")
                        cleanup()
                        deferred.complete(Result.failure(Exception(error)))
                    }
                })
            } catch (e: Exception) {
                Log.e("VpnConnectionService", "Connect failed", e)
                cleanup()
                deferred.complete(Result.failure(e))
            }
        }
        return deferred.await()
    }

    fun disconnect() {
        scope.launch {
            cleanup()
        }
    }

    private fun cleanup() {
        try {
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