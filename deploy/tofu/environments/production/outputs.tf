output "control_ipv4" {
  value = module.cluster.control_ipv4
}

output "load_balancer_ipv4" {
  value = module.cluster.load_balancer_ipv4
}

output "load_balancer_ipv6" {
  value = module.cluster.load_balancer_ipv6
}

output "node_names" {
  value = module.cluster.node_names
}

output "k3s_version" {
  value = module.cluster.k3s_version
}
