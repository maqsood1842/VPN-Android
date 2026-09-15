package com.vpn.core.network

import com.vpn.core.data.Models.PeerConfig
import com.vpn.core.data.Models.Server
import retrofit2.Response
import retrofit2.http.*

interface ApiService {
    @POST("auth/signup")
    suspend fun signup(@Body request: SignupRequest): Response<AuthResponse>

    @POST("auth/login")
    suspend fun login(@Body request: LoginRequest): Response<AuthResponse>

    @POST("auth/refresh")
    suspend fun refresh(@Body request: RefreshRequest): Response<TokenPair>

    @GET("servers")
    suspend fun getServers(@Header("Authorization") authHeader: String): Response<List<Server>>

    @GET("wireguard/config/{server_id}")
    suspend fun getWireGuardConfig(
        @Header("Authorization") authHeader: String,
        @Path("server_id") serverId: String
    ): Response<PeerConfig>

    @GET("wireguard/peers")
    suspend fun getPeers(@Header("Authorization") authHeader: String): Response<List<PeerConfig>>
}

data class SignupRequest(
    val email: String,
    val password: String
)

data class LoginRequest(
    val email: String,
    val password: String
)

data class RefreshRequest(
    val refresh_token: String
)

data class AuthResponse(
    val user: User,
    val tokens: TokenPair
)

data class User(
    val id: String,
    val email: String,
    val created_at: String,
    val updated_at: String
)

data class TokenPair(
    val access_token: String,
    val refresh_token: String
)