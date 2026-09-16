package com.vpn.core.network

import com.vpn.core.data.TokenStore
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class AuthRepository @Inject constructor(
    private val apiClient: ApiClient,
    private val tokenStore: TokenStore
) {
    suspend fun signup(email: String, password: String): Result<TokenPair> {
        val response = apiClient.apiService.signup(SignupRequest(email, password))
        return handleAuthResponse(response)
    }

    suspend fun login(email: String, password: String): Result<TokenPair> {
        val response = apiClient.apiService.login(LoginRequest(email, password))
        return handleAuthResponse(response)
    }

    suspend fun refreshToken(): Result<TokenPair> {
        val refreshToken = tokenStore.getRefreshToken() ?: return Result.failure(Exception("No refresh token"))
        val response = apiClient.apiService.refresh(RefreshRequest(refreshToken))
        return handleTokenResponse(response)
    }

    private fun handleAuthResponse(response: retrofit2.Response<AuthResponse>): Result<TokenPair> {
        if (response.isSuccessful) {
            response.body()?.tokens?.let { tokens ->
                tokenStore.saveTokens(
                    tokens.access_token,
                    tokens.refresh_token,
                    response.body()!!.user.id,
                    response.body()!!.user.email
                )
                return Result.success(tokens)
            }
        }
        return Result.failure(Exception(response.errorBody()?.string() ?: "Auth failed"))
    }

    private fun handleTokenResponse(response: retrofit2.Response<TokenPair>): Result<TokenPair> {
        if (response.isSuccessful) {
            response.body()?.let { tokens ->
                tokenStore.saveTokens(
                    tokens.access_token,
                    tokens.refresh_token,
                    tokenStore.getUserId() ?: "",
                    tokenStore.getUserEmail() ?: ""
                )
                return Result.success(tokens)
            }
        }
        return Result.failure(Exception(response.errorBody()?.string() ?: "Token refresh failed"))
    }

    fun getAuthHeader(): String? {
        val token = tokenStore.getAccessToken()
        return token?.let { "Bearer $it" }
    }

    fun getUserEmail(): String? = tokenStore.getUserEmail()

    fun getUserId(): String? = tokenStore.getUserId()

    fun isLoggedIn(): Boolean = tokenStore.hasValidTokens()

    fun logout() {
        tokenStore.clearTokens()
    }
}