package com.vpn.app.ui.auth

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.vpn.core.network.AuthRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class SignupViewModel @Inject constructor(
    private val authRepository: AuthRepository
) : ViewModel() {

    fun signup(email: String, password: String, callback: (Boolean) -> Unit) {
        viewModelScope.launch {
            val result = authRepository.signup(email, password)
            result.onSuccess {
                callback(true)
            }.onFailure { error ->
                callback(false)
            }
        }
    }
}