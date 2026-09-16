package com.vpn.app.ui.settings

import android.os.Bundle
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import androidx.fragment.app.Fragment
import androidx.fragment.app.viewModels
import androidx.lifecycle.Observer
import com.vpn.app.R
import com.vpn.app.databinding.FragmentSettingsBinding
import com.vpn.core.vpn.VpnConnectionService
import dagger.hilt.android.AndroidEntryPoint
import javax.inject.Inject

@AndroidEntryPoint
class SettingsFragment : Fragment(R.layout.fragment_settings) {

    @Inject
    lateinit var vpnService: VpnConnectionService

    private var _binding: FragmentSettingsBinding? = null
    private val binding get() = _binding!!

    private val viewModel: SettingsViewModel by viewModels()

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        _binding = FragmentSettingsBinding.bind(view)

        binding.killSwitchToggle.setOnCheckedChangeListener { _, isChecked ->
            vpnService.setKillSwitchEnabled(isChecked)
            viewModel.setKillSwitchEnabled(isChecked)
        }

        observeState()
    }

    private fun observeState() {
        viewModel.killSwitchEnabled.observe(viewLifecycleOwner, Observer { enabled ->
            binding.killSwitchToggle.isChecked = enabled
        })
    }

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }
}