package com.vpn.app.ui.account

import android.os.Bundle
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import androidx.fragment.app.Fragment
import androidx.fragment.app.viewModels
import androidx.navigation.fragment.findNavController
import com.vpn.app.R
import com.vpn.app.databinding.FragmentAccountBinding
import dagger.hilt.android.AndroidEntryPoint
import javax.inject.Inject

@AndroidEntryPoint
class AccountFragment : Fragment(R.layout.fragment_account) {

    @Inject
    lateinit var authRepository: com.vpn.core.network.AuthRepository

    private var _binding: FragmentAccountBinding? = null
    private val binding get() = _binding!!

    private val viewModel: AccountViewModel by viewModels()

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        _binding = FragmentAccountBinding.bind(view)

        binding.logoutButton.setOnClickListener {
            authRepository.logout()
            findNavController().navigate(R.id.loginFragment) {
                popUpTo(R.id.loginFragment) { inclusive = true }
            }
        }

        binding.manageSubscriptionButton.setOnClickListener {
            // Navigate to subscription management
        }

        observeState()
    }

    private fun observeState() {
        viewModel.userEmail.observe(viewLifecycleOwner) { email ->
            binding.accountEmail.text = email ?: "Not signed in"
        }

        viewModel.subscriptionStatus.observe(viewLifecycleOwner) { status ->
            binding.subscriptionStatus.text = status
        }
    }

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }
}