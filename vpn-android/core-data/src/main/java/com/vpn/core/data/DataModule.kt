package com.vpn.core.data

import android.content.Context
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
object DataModule {
    @Provides
    @Singleton
    fun provideTokenStore(@dagger.hilt.android.qualifiers.ApplicationContext context: Context): TokenStore {
        return TokenStore(context)
    }
}