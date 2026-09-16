package com.vpn.app.ui.settings

import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData
import androidx.lifecycle.ViewModel
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject

@HiltViewModel
class SettingsViewModel @Inject constructor() : ViewModel() {

    val killSwitchEnabled: MutableLiveData<Boolean> = MutableLiveData(true)
    val dnsLeakProtectionEnabled: MutableLiveData<Boolean> = MutableLiveData(true)

    fun setKillSwitchEnabled(enabled: Boolean) {
        killSwitchEnabled.value = enabled
    }

    fun setDnsLeakProtectionEnabled(enabled: Boolean) {
        dnsLeakProtectionEnabled.value = enabled
    }
}