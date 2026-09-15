package com.vpn.core.data

import kotlinx.serialization.Serializable

@Serializable
data class PeerConfig(
    val interface: InterfaceConfig,
    val peer: PeerConfigData
)

@Serializable
data class InterfaceConfig(
    val privateKey: String,
    val address: List<String>,
    val dns: List<String>
)

@Serializable
data class PeerConfigData(
    val publicKey: String,
    val presharedKey: String,
    val allowedIPs: List<String>,
    val endpoint: String,
    val persistentKeepalive: Int
)

@Serializable
data class Server(
    val id: String,
    val name: String,
    val region: String,
    val host: String,
    val port: Int,
    val publicKey: String,
    val endpointIp: String?,
    val status: String,
    val loadPercent: Int,
    val maxClients: Int
)