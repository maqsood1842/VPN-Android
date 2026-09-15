package com.vpn.app.ui.servers

import android.os.Bundle
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import androidx.fragment.app.Fragment
import androidx.fragment.app.viewModels
import androidx.lifecycle.Observer
import androidx.navigation.fragment.findNavController
import androidx.recyclerview.widget.LinearLayoutManager
import com.vpn.app.R
import com.vpn.app.databinding.FragmentServersBinding
import com.vpn.core.data.Models.Server
import dagger.hilt.android.AndroidEntryPoint
import javax.inject.Inject

@AndroidEntryPoint
class ServersFragment : Fragment(R.layout.fragment_servers) {

    @Inject
    lateinit var wireGuardRepository: com.vpn.core.network.WireGuardRepository

    private var _binding: FragmentServersBinding? = null
    private val binding get() = _binding!!

    private val viewModel: ServersViewModel by viewModels()
    private val adapter = ServerAdapter()

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        _binding = FragmentServersBinding.bind(view)

        binding.serversRecyclerView.layoutManager = LinearLayoutManager(requireContext())
        binding.serversRecyclerView.adapter = adapter

        adapter.onServerClick = { server ->
            viewModel.selectServer(server)
            findNavController().navigateUp()
        }

        viewModel.servers.observe(viewLifecycleOwner, Observer { servers ->
            adapter.submitList(servers)
        })

        viewModel.loadServers()
    }

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }
}

class ServerAdapter : androidx.recyclerview.widget.ListAdapter<Server, ServerAdapter.ServerViewHolder>(
    androidx.diffutil.DiffUtil.ItemCallback<Server> { old, new ->
        old.id == new.id
    }
) {
    var onServerClick: ((Server) -> Unit)? = null

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): ServerViewHolder {
        val view = LayoutInflater.from(parent.context)
            .inflate(R.layout.item_server, parent, false)
        return ServerViewHolder(view)
    }

    override fun onBindViewHolder(holder: ServerViewHolder, position: Int) {
        holder.bind(getItem(position))
    }

    inner class ServerViewHolder(itemView: View) : androidx.recyclerview.widget.RecyclerView.ViewHolder(itemView) {
        fun bind(server: Server) {
            itemView.findViewById<android.widget.TextView>(R.id.server_name).text = server.name
            itemView.findViewById<android.widget.TextView>(R.id.server_region).text = server.region
            itemView.findViewById<android.widget.TextView>(R.id.server_load).text = "${server.loadPercent}% load"
            itemView.setOnClickListener { onServerClick?.invoke(server) }
        }
    }
}