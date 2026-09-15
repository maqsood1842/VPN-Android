package com.vpn.core.network

import com.vpn.core.data.Models.PeerConfig
import com.vpn.core.data.Models.Server
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class WireGuardRepository @Inject constructor(
    private val apiClient: ApiClient,
    private val authRepository: AuthRepository
) {
    suspend fun getServers(): Result<List<Server>> {
        val authHeader = authRepository.getAuthHeader() ?: return Result.failure(Exception("Not authenticated"))
        val response = apiClient.apiService.getServers(authHeader)
        return if (response.isSuccessful) {
            Result.success(response.body() ?: emptyList())
        } else {
            Result.failure(Exception(response.errorBody()?.string() ?: "Failed to fetch servers"))
        }
    }

    suspend fun getConfig(serverId: String): Result<PeerConfig> {
        val authHeader = authRepository.getAuthHeader() ?: return Result.failure(Exception("Not authenticated"))
        val response = apiClient.apiService.getWireGuardConfig(authHeader, serverId)
        return if (response.isSuccessful) {
            Result.success(response.body()!!)
        } else {
            Result.failure(Exception(response.errorBody()?.string() ?: "Failed to fetch config"))
        }
    }
}