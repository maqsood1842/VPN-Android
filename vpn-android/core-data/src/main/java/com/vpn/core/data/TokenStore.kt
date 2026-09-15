package com.vpn.core.data

import android.content.Context
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.rx.preferenceDataStore
import androidx.security.crypto.EncryptedSharedPreferences
import androidx.security.crypto.MasterKey
import io.reactivex.rxjava3.core.Completable
import io.reactivex.rxjava3.core.Single
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class TokenStore @Inject constructor(
    private val context: Context
) {
    private val masterKey by lazy {
        MasterKey.Builder(context)
            .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
            .build()
    }

    private val encryptedPrefs by lazy {
        EncryptedSharedPreferences.create(
            context,
            "secure_token_store",
            masterKey,
            EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
            EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM
        )
    }

    private val datastore = context.preferenceDataStore("token_datastore")

    private val ACCESS_TOKEN_KEY = stringPreferencesKey("access_token")
    private val REFRESH_TOKEN_KEY = stringPreferencesKey("refresh_token")
    private val USER_EMAIL_KEY = stringPreferencesKey("user_email")
    private val USER_ID_KEY = stringPreferencesKey("user_id")

    suspend fun saveTokens(accessToken: String, refreshToken: String, userId: String, email: String) {
        encryptedPrefs.edit()
            .putString("access_token", accessToken)
            .putString("refresh_token", refreshToken)
            .putString("user_id", userId)
            .putString("user_email", email)
            .apply()
    }

    fun getAccessToken(): String? = encryptedPrefs.getString("access_token", null)
    fun getRefreshToken(): String? = encryptedPrefs.getString("refresh_token", null)
    fun getUserId(): String? = encryptedPrefs.getString("user_id", null)
    fun getUserEmail(): String? = encryptedPrefs.getString("user_email", null)

    fun clearTokens() {
        encryptedPrefs.edit().clear().apply()
    }

    fun hasValidTokens(): Boolean {
        return getAccessToken() != null && getRefreshToken() != null
    }
}