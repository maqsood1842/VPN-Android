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
  value       = digitalocean_droplet.vpn_node.provisioner[0].connection.host
  sensitive   = true
}

output "dns_record" {
  description = "DNS record name for the VPN endpoint"
  value       = var.node_name
}