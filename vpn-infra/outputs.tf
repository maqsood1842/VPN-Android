output "droplet_ip" {
  description = "Public IP of the VPN droplet"
  value       = digitalocean_droplet.vpn_node.ipv4_address
}

output "droplet_id" {
  description = "ID of the VPN droplet"
  value       = digitalocean_droplet.vpn_node.id
}

output "wireguard_public_key" {
  description = "WireGuard server public key"
  value       = module.vpn_node.wireguard_public_key
  sensitive   = true
}

output "wireguard_endpoint" {
  description = "WireGuard endpoint (domain:port)"
  value       = "${module.vpn_node.dns_record}.${var.domain_name}:${var.wireguard_port}"
}