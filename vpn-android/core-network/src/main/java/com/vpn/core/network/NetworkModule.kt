package com.vpn.core.network

import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
object NetworkModule {
    @Provides
    @Singleton
    fun provideApiClient(): ApiClient {
        return ApiClient()
    }

    @Provides
    @Singleton
    fun provideAuthRepository(apiClient: ApiClient, tokenStore: com.vpn.core.data.TokenStore): AuthRepository {
        return AuthRepository(apiClient, tokenStore)
    }

    @Provides
    @Singleton
    fun provideWireGuardRepository(apiClient: ApiClient, authRepository: AuthRepository): WireGuardRepository {
        return WireGuardRepository(apiClient, authRepository)
    }
}