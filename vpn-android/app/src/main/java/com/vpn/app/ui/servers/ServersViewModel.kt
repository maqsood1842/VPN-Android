package com.vpn.app.ui.servers

import androidx.lifecycle.LiveData
import androidx.lifecycle.ViewModel
import androidx.lifecycle.asLiveData
import androidx.lifecycle.viewModelScope
import com.vpn.core.data.Models.Server
import com.vpn.core.network.WireGuardRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class ServersViewModel @Inject constructor(
    private val wireGuardRepository: WireGuardRepository
) : ViewModel() {

    private val _servers = MutableStateFlow<List<Server>>(emptyList())
    val servers: LiveData<List<Server>> = _servers.asLiveData(viewModelScope.coroutineContext)

    var onServerSelected: ((Server) -> Unit)? = null

    fun loadServers() {
        viewModelScope.launch {
            val result = wireGuardRepository.getServers()
            result.onSuccess { serverList ->
                _servers.value = serverList
            }.onFailure { error ->
                // Handle error - could show in UI
            }
        }
    }

    fun selectServer(server: Server) {
        onServerSelected?.invoke(server)
    }
}