module "vpn_node_nyc" {
  source             = "../../modules/vpn-node"
  region             = "nyc1"
  droplet_size       = "s-1vcpu-1gb"
  ssh_public_key     = var.ssh_public_key
  ssh_private_key    = var.ssh_private_key
  wireguard_port     = 51820
  domain_name        = var.domain_name
  node_name          = "nyc1"
  cloudflare_zone_id = var.cloudflare_zone_id
}

module "vpn_node_lon" {
  source             = "../../modules/vpn-node"
  region             = "lon1"
  droplet_size       = "s-1vcpu-1gb"
  ssh_public_key     = var.ssh_public_key
  ssh_private_key    = var.ssh_private_key
  wireguard_port     = 51820
  domain_name        = var.domain_name
  node_name          = "lon1"
  cloudflare_zone_id = var.cloudflare_zone_id
}

module "vpn_node_sfo" {
  source             = "../../modules/vpn-node"
  region             = "sfo3"
  droplet_size       = "s-1vcpu-1gb"
  ssh_public_key     = var.ssh_public_key
  ssh_private_key    = var.ssh_private_key
  wireguard_port     = 51820
  domain_name        = var.domain_name
  node_name          = "sfo3"
  cloudflare_zone_id = var.cloudflare_zone_id
}