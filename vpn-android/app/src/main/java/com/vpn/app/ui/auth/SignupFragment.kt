package com.vpn.app.ui.auth

import android.os.Bundle
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import androidx.fragment.app.Fragment
import androidx.fragment.app.viewModels
import androidx.navigation.fragment.findNavController
import com.vpn.app.R
import com.vpn.app.databinding.FragmentSignupBinding
import dagger.hilt.android.AndroidEntryPoint
import javax.inject.Inject

@AndroidEntryPoint
class SignupFragment : Fragment(R.layout.fragment_signup) {

    @Inject
    lateinit var authRepository: com.vpn.core.network.AuthRepository

    private var _binding: FragmentSignupBinding? = null
    private val binding get() = _binding!!

    private val viewModel: SignupViewModel by viewModels()

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        _binding = FragmentSignupBinding.bind(view)

        binding.signupButton.setOnClickListener { attemptSignup() }
        binding.loginLink.setOnClickListener {
            findNavController().navigate(R.id.loginFragment)
        }
    }

    private fun attemptSignup() {
        val email = binding.emailInput.text.toString().trim()
        val password = binding.passwordInput.text.toString()
        val confirmPassword = binding.confirmPasswordInput.text.toString()

        if (email.isEmpty() || password.isEmpty() || password != confirmPassword) {
            return
        }

        binding.signupButton.isEnabled = false
        binding.signupButton.text = "Creating account..."

        viewModel.signup(email, password) { success ->
            binding.signupButton.isEnabled = true
            binding.signupButton.text = getString(R.string.signup)
            if (success) {
                findNavController().navigate(R.id.homeFragment) {
                    popUpTo(R.id.signupFragment) { inclusive = true }
                }
            }
        }
    }

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }
}