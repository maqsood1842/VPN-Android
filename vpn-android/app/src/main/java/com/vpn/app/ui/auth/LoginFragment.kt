package com.vpn.app.ui.auth

import android.os.Bundle
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import androidx.fragment.app.Fragment
import androidx.fragment.app.viewModels
import androidx.navigation.fragment.findNavController
import com.vpn.app.R
import com.vpn.app.databinding.FragmentLoginBinding
import dagger.hilt.android.AndroidEntryPoint
import javax.inject.Inject

@AndroidEntryPoint
class LoginFragment : Fragment(R.layout.fragment_login) {

    @Inject
    lateinit var authRepository: com.vpn.core.network.AuthRepository

    private var _binding: FragmentLoginBinding? = null
    private val binding get() = _binding!!

    private val viewModel: LoginViewModel by viewModels()

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        _binding = FragmentLoginBinding.bind(view)

        binding.loginButton.setOnClickListener { attemptLogin() }
        binding.signupLink.setOnClickListener {
            findNavController().navigate(R.id.signupFragment)
        }
    }

    private fun attemptLogin() {
        val email = binding.emailInput.text.toString().trim()
        val password = binding.passwordInput.text.toString()

        if (email.isEmpty() || password.isEmpty()) {
            return
        }

        binding.loginButton.isEnabled = false
        binding.loginButton.text = "Signing in..."

        viewModel.login(email, password) { success ->
            binding.loginButton.isEnabled = true
            binding.loginButton.text = getString(R.string.login)
            if (success) {
                findNavController().navigate(R.id.homeFragment) {
                    popUpTo(R.id.loginFragment) { inclusive = true }
                }
            }
        }
    }

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }
}