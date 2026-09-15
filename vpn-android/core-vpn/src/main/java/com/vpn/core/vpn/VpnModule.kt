package com.vpn.core.vpn

import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
object VpnModule {
    @Provides
    @Singleton
    fun provideVpnConnectionService(@dagger.hilt.android.qualifiers.ApplicationContext context: android.content.Context): VpnConnectionService {
        return VpnConnectionService(context)
    }
}