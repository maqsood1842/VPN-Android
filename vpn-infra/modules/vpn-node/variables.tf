variable "region" {
  type        = string
  description = "DigitalOcean region"
}

variable "droplet_size" {
  type        = string
  description = "Droplet size slug"
}

variable "ssh_public_key" {
  type        = string
  description = "SSH public key"
}

variable "ssh_private_key" {
  type        = string
  description = "SSH private key"
  sensitive   = true
}

variable "wireguard_port" {
  type        = number
  description = "WireGuard listen port"
  default     = 51820
}

variable "domain_name" {
  type        = string
  description = "Base domain name"
}

variable "node_name" {
  type        = string
  description = "Unique node name"
}

variable "cloudflare_zone_id" {
  type        = string
  description = "Cloudflare zone ID"
}