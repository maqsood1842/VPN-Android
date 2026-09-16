package com.vpn.app.ui.account

import android.os.Bundle
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.Toast
import androidx.fragment.app.Fragment
import androidx.fragment.app.viewModels
import androidx.navigation.fragment.findNavController
import com.vpn.app.R
import com.vpn.app.databinding.FragmentAccountBinding
import com.vpn.core.network.BillingService
import dagger.hilt.android.AndroidEntryPoint
import javax.inject.Inject

@AndroidEntryPoint
class AccountFragment : Fragment(R.layout.fragment_account) {

    @Inject
    lateinit var authRepository: com.vpn.core.network.AuthRepository

    @Inject
    lateinit var billingService: BillingService

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
            launchSubscriptionPurchase()
        }

        binding.subscribeButton.setOnClickListener {
            launchSubscriptionPurchase()
        }

        observeState()
    }

    private fun launchSubscriptionPurchase() {
        val productId = "vpn_monthly_subscription" // Replace with your actual product ID from Play Console
        
        billingService.launchBillingFlow(requireActivity(), productId) { billingResult ->
            requireActivity().runOnUiThread {
                when (billingResult.responseCode) {
                    com.android.billingclient.api.BillingClient.BillingResponseCode.OK -> {
                        // Purchase flow launched successfully
                    }
                    com.android.billingclient.api.BillingClient.BillingResponseCode.USER_CANCELED -> {
                        Toast.makeText(requireContext(), "Purchase cancelled", Toast.LENGTH_SHORT).show()
                    }
                    com.android.billingclient.api.BillingClient.BillingResponseCode.ITEM_ALREADY_OWNED -> {
                        Toast.makeText(requireContext(), "Already subscribed", Toast.LENGTH_SHORT).show()
                    }
                    else -> {
                        Toast.makeText(requireContext(), "Purchase failed: ${billingResult.debugMessage}", Toast.LENGTH_SHORT).show()
                    }
                }
            }
        }
    }

    private fun observeState() {
        viewModel.userEmail.observe(viewLifecycleOwner) { email ->
            binding.accountEmail.text = email ?: "Not signed in"
        }

        viewModel.subscriptionStatus.observe(viewLifecycleOwner) { status ->
            binding.subscriptionStatus.text = status
            val isSubscribed = status.contains("Active", ignoreCase = true)
            binding.subscribeButton.visibility = if (isSubscribed) View.GONE else View.VISIBLE
            binding.manageSubscriptionButton.visibility = if (isSubscribed) View.VISIBLE else View.GONE
        }
    }

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }
}