package com.vpn.app.ui.home

import android.os.Bundle
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import androidx.fragment.app.Fragment
import androidx.fragment.app.viewModels
import androidx.navigation.fragment.findNavController
import com.vpn.app.R
import com.vpn.app.databinding.FragmentHomeBinding
import com.vpn.core.vpn.VpnConnectionService
import dagger.hilt.android.AndroidEntryPoint
import javax.inject.Inject

@AndroidEntryPoint
class HomeFragment : Fragment(R.layout.fragment_home) {

    @Inject
    lateinit var vpnService: VpnConnectionService

    private var _binding: FragmentHomeBinding? = null
    private val binding get() = _binding!!

    private val viewModel: HomeViewModel by viewModels()

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        _binding = FragmentHomeBinding.bind(view)

        binding.connectButton.setOnClickListener { toggleConnection() }
        binding.selectServerButton.setOnClickListener {
            findNavController().navigate(R.id.serversFragment)
        }

        observeState()
    }

    private fun observeState() {
        viewModel.connectionState.observe(viewLifecycleOwner) { state ->
            when (state) {
                is ConnectionState.Connected -> {
                    binding.vpnStatusIcon.setImageResource(R.drawable.ic_vpn_connected)
                    binding.vpnStatusText.text = getString(R.string.connected)
                    binding.connectButton.text = getString(R.string.disconnect)
                    binding.currentServerText.text = "Connected to ${viewModel.selectedServer?.name ?: "VPN"}"
                }
                is ConnectionState.Connecting -> {
                    binding.vpnStatusIcon.setImageResource(R.drawable.ic_vpn_connecting)
                    binding.vpnStatusText.text = getString(R.string.connecting)
                    binding.connectButton.isEnabled = false
                }
                is ConnectionState.Disconnected -> {
                    binding.vpnStatusIcon.setImageResource(R.drawable.ic_vpn_disconnected)
                    binding.vpnStatusText.text = getString(R.string.disconnected)
                    binding.connectButton.text = getString(R.string.connect)
                    binding.connectButton.isEnabled = true
                    binding.currentServerText.text = viewModel.selectedServer?.name?.let { "Selected: $it" } ?: "No server selected"
                }
                is ConnectionState.Error -> {
                    binding.vpnStatusIcon.setImageResource(R.drawable.ic_vpn_disconnected)
                    binding.vpnStatusText.text = "Error: ${state.message}"
                    binding.connectButton.isEnabled = true
                    binding.connectButton.text = getString(R.string.connect)
                }
            }
        }
    }

    private fun toggleConnection() {
        when (viewModel.connectionState.value) {
            is ConnectionState.Connected -> {
                viewModel.disconnect()
            }
            is ConnectionState.Disconnected -> {
                viewModel.connect()
            }
            else -> {}
        }
    }

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }
}

sealed interface ConnectionState {
    data class Connected(val serverName: String) : ConnectionState
    object Connecting : ConnectionState
    object Disconnected : ConnectionState
    data class Error(val message: String) : ConnectionState
}