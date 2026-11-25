output "control_ipv4" {
  value = hcloud_primary_ip.control.ip_address
}

output "load_balancer_ipv4" {
  value = hcloud_load_balancer.game.ipv4
}

output "load_balancer_ipv6" {
  value = hcloud_load_balancer.game.ipv6
}

output "node_names" {
  value = merge(
    { control = hcloud_server.control.name },
    { for key, server in hcloud_server.agent : key => server.name },
  )
}

output "k3s_version" {
  value = local.k3s_version
}
