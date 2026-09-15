resource "digitalocean_droplet" "vpn_node" {
  name       = "vpn-${var.node_name}"
  region     = var.region
  size       = var.droplet_size
  image      = "ubuntu-22-04-x64"
  ssh_keys   = [data.digitalocean_ssh_key.ssh_key.fingerprint]

  connection {
    type        = "ssh"
    user        = "root"
    private_key = var.ssh_private_key
    host        = self.ipv4_address
    timeout     = "5m"
  }

  provisioner "remote-exec" {
    inline = [
      "export DEBIAN_FRONTEND=noninteractive",
      "apt-get update",
      "apt-get install -y wireguard wireguard-tools iptables-persistent qrencode",
      "mkdir -p /etc/wireguard",
      "wg genkey | tee /etc/wireguard/private.key | wg pubkey > /etc/wireguard/public.key",
      "chmod 600 /etc/wireguard/private.key",
    ]
  }

  provisioner "remote-exec" {
    inline = [
      "cat > /etc/wireguard/wg0.conf << 'EOF'\n[Interface]\nAddress = 10.0.0.1/24, fd00::1/64\nListenPort = ${var.wireguard_port}\nPrivateKey = $(cat /etc/wireguard/private.key)\nPostUp = iptables -A FORWARD -i %i -j ACCEPT; iptables -A FORWARD -o %i -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE\nPostDown = iptables -D FORWARD -i %i -j ACCEPT; iptables -D FORWARD -o %i -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE\nEOF",
      "systemctl enable wg-quick@wg0",
      "systemctl start wg-quick@wg0",
      "sysctl -w net.ipv4.ip_forward=1",
      "sysctl -w net.ipv6.conf.all.forwarding=1",
      "echo 'net.ipv4.ip_forward=1' >> /etc/sysctl.conf",
      "echo 'net.ipv6.conf.all.forwarding=1' >> /etc/sysctl.conf",
    ]
  }
}

resource "digitalocean_firewall" "vpn_firewall" {
  name = "vpn-firewall-${var.node_name}"
  droplet_ids = [digitalocean_droplet.vpn_node.id]

  inbound_rule {
    protocol         = "udp"
    port_range       = "${var.wireguard_port}"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }

  inbound_rule {
    protocol         = "tcp"
    port_range       = "22"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }

  outbound_rule {
    protocol              = "tcp"
    port_range            = "1-65535"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }

  outbound_rule {
    protocol              = "udp"
    port_range            = "1-65535"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }

  outbound_rule {
    protocol              = "icmp"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }
}

data "digitalocean_ssh_key" "ssh_key" {
  name = "vpn-deploy-key"
}

resource "cloudflare_record" "vpn_dns" {
  zone_id = var.cloudflare_zone_id
  name    = var.node_name
  value   = digitalocean_droplet.vpn_node.ipv4_address
  type    = "A"
  ttl     = 300
  proxied = false
}