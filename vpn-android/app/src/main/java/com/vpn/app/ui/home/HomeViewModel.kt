package com.vpn.app.ui.home

import androidx.lifecycle.LiveData
import androidx.lifecycle.ViewModel
import androidx.lifecycle.asLiveData
import androidx.lifecycle.viewModelScope
import com.vpn.core.data.Models.Server
import com.vpn.core.network.WireGuardRepository
import com.vpn.core.vpn.VpnConnectionService
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class HomeViewModel @Inject constructor(
    private val wireGuardRepository: WireGuardRepository,
    private val vpnService: VpnConnectionService,
    private val authRepository: com.vpn.core.network.AuthRepository
) : ViewModel() {

    private val _connectionState = MutableStateFlow<ConnectionState>(ConnectionState.Disconnected)
    val connectionState: LiveData<ConnectionState> = _connectionState.asLiveData(viewModelScope.coroutineContext)

    var selectedServer: Server? = null
        private set

    fun connect() {
        selectedServer?.let { server ->
            viewModelScope.launch {
                _connectionState.value = ConnectionState.Connecting
                val result = wireGuardRepository.getConfig(server.id)
                result.onSuccess { config ->
                    vpnService.connect(config)
                        .onSuccess {
                            _connectionState.value = ConnectionState.Connected(server.name)
                        }
                        .onFailure { error ->
                            _connectionState.value = ConnectionState.Error(error.message ?: "Connection failed")
                        }
                }.onFailure { error ->
                    _connectionState.value = ConnectionState.Error(error.message ?: "Failed to get config")
                }
            }
        }
    }

    fun disconnect() {
        viewModelScope.launch {
            vpnService.disconnect()
            _connectionState.value = ConnectionState.Disconnected
        }
    }

    fun setSelectedServer(server: Server) {
        selectedServer = server
        _connectionState.value = ConnectionState.Disconnected
    }
}