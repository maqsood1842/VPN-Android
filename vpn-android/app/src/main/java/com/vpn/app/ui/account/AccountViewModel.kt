package com.vpn.app.ui.account

import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData
import androidx.lifecycle.ViewModel
import com.vpn.core.network.AuthRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject

@HiltViewModel
class AccountViewModel @Inject constructor(
    private val authRepository: AuthRepository
) : ViewModel() {

    val userEmail: MutableLiveData<String?> = MutableLiveData()
    val subscriptionStatus: MutableLiveData<String> = MutableLiveData("No active subscription")

    init {
        loadUserInfo()
    }

    private fun loadUserInfo() {
        userEmail.value = authRepository.getUserEmail()
        // Subscription status would be loaded from backend
    }
}