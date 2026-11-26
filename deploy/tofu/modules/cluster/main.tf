locals {
  name_prefix = "hd-${var.environment}"

  k3s_version               = "v1.37.2+k3s1"
  k3s_install_script_sha256 = "ed01f89fd977bf20ac1516bbebf8370bf3ddbaa55dac8aba610956a4c78cc00b"
  hcloud_csi_chart_version  = "2.24.1"

  network_range      = "10.0.0.0/16"
  subnet_range       = "10.0.1.0/24"
  node_subnet_prefix = "10.0.1."
  control_private_ip = "10.0.1.10"
  network_zone       = "eu-central"

  listen_backlog        = 65535
  listen_backlog_sysctl = "net.core.somaxconn"
  kubelet_args          = ["allowed-unsafe-sysctls=${local.listen_backlog_sysctl}"]
  game_sysctl_file = {
    path        = "/etc/sysctl.d/90-hypedomino-listen.conf"
    permissions = "0644"
    content     = "${local.listen_backlog_sysctl} = ${local.listen_backlog}\n"
  }

  common_labels = {
    "hypedomino.io/environment" = var.environment
    "managed-by"                = "opentofu"
  }

  agent_nodes = merge(
    { db = merge(var.db_node, { role = "db" }) },
    { for name, node in var.game_nodes : name => merge(node, { role = "game" }) },
  )

  install_script = templatefile("${path.module}/k3s-install.sh.tftpl", {
    node_subnet_prefix = local.node_subnet_prefix
    k3s_version        = local.k3s_version
    k3s_version_url    = replace(local.k3s_version, "+", "%2B")
    install_sha256     = local.k3s_install_script_sha256
  })

  hcloud_csi_manifest = yamlencode({
    apiVersion = "helm.cattle.io/v1"
    kind       = "HelmChart"
    metadata = {
      name      = "hcloud-csi"
      namespace = "kube-system"
    }
    spec = {
      repo            = "https://charts.hetzner.cloud"
      chart           = "hcloud-csi"
      version         = local.hcloud_csi_chart_version
      targetNamespace = "kube-system"
      valuesContent = yamlencode({
        storageClasses = [{
          name                = "hcloud-volumes"
          defaultStorageClass = true
          reclaimPolicy       = "Retain"
          annotations         = {}
          extraParameters = {
            "csi.storage.k8s.io/node-publish-secret-name"      = "encryption-secret"
            "csi.storage.k8s.io/node-publish-secret-namespace" = "kube-system"
          }
        }]
      })
    }
  })

  control_k3s_config = yamlencode({
    token                   = var.k3s_token
    "advertise-address"     = local.control_private_ip
    "tls-san"               = [hcloud_primary_ip.control.ip_address, local.control_private_ip]
    "node-label"            = ["hypedomino.io/role=control"]
    "kubelet-arg"           = local.kubelet_args
    disable                 = ["traefik", "servicelb", "local-storage"]
    "secrets-encryption"    = true
    "write-kubeconfig-mode" = "0600"
  })
}

resource "hcloud_network" "cluster" {
  name              = local.name_prefix
  ip_range          = local.network_range
  labels            = local.common_labels
  delete_protection = var.protected
}

resource "hcloud_network_subnet" "nodes" {
  network_id   = hcloud_network.cluster.id
  type         = "cloud"
  network_zone = local.network_zone
  ip_range     = local.subnet_range
}

resource "hcloud_ssh_key" "operators" {
  for_each   = var.ssh_public_keys
  name       = "${local.name_prefix}-${each.key}"
  public_key = each.value
  labels     = local.common_labels
}

resource "hcloud_firewall" "nodes" {
  name   = local.name_prefix
  labels = local.common_labels

  rule {
    description = "icmp"
    direction   = "in"
    protocol    = "icmp"
    source_ips  = ["0.0.0.0/0", "::/0"]
  }

  dynamic "rule" {
    for_each = length(var.ssh_allowed_cidrs) > 0 ? [1] : []
    content {
      description = "ssh"
      direction   = "in"
      protocol    = "tcp"
      port        = "22"
      source_ips  = var.ssh_allowed_cidrs
    }
  }

  dynamic "rule" {
    for_each = length(var.api_allowed_cidrs) > 0 ? [1] : []
    content {
      description = "kubernetes-api"
      direction   = "in"
      protocol    = "tcp"
      port        = "6443"
      source_ips  = var.api_allowed_cidrs
    }
  }
}

resource "hcloud_primary_ip" "control" {
  name              = "${local.name_prefix}-control"
  type              = "ipv4"
  location          = var.control_node.location
  auto_delete       = false
  delete_protection = var.protected
  labels            = local.common_labels
}

resource "hcloud_server" "control" {
  name               = "${local.name_prefix}-control"
  server_type        = var.control_node.server_type
  location           = var.control_node.location
  image              = var.image
  ssh_keys           = [for key in hcloud_ssh_key.operators : key.id]
  firewall_ids       = [hcloud_firewall.nodes.id]
  delete_protection  = var.protected
  rebuild_protection = var.protected
  labels             = merge(local.common_labels, { "hypedomino.io/role" = "control" })

  user_data = "#cloud-config\n${yamlencode({
    write_files = [
      { path = "/etc/rancher/k3s/config.yaml", permissions = "0600", content = local.control_k3s_config },
      { path = "/var/lib/rancher/k3s/server/manifests/hcloud-csi.yaml", permissions = "0600", content = local.hcloud_csi_manifest },
      { path = "/usr/local/sbin/hypedomino-install-k3s", permissions = "0700", content = local.install_script },
    ]
    runcmd = [["/usr/local/sbin/hypedomino-install-k3s", "server"]]
  })}"

  public_net {
    ipv4_enabled = true
    ipv4         = hcloud_primary_ip.control.id
    ipv6_enabled = true
  }

  network {
    network_id = hcloud_network.cluster.id
    ip         = local.control_private_ip
  }

  lifecycle {
    ignore_changes = [user_data, image, ssh_keys]
  }

  depends_on = [hcloud_network_subnet.nodes]
}

resource "hcloud_server" "agent" {
  for_each = local.agent_nodes

  name               = "${local.name_prefix}-${each.key}"
  server_type        = each.value.server_type
  location           = each.value.location
  image              = var.image
  ssh_keys           = [for key in hcloud_ssh_key.operators : key.id]
  firewall_ids       = [hcloud_firewall.nodes.id]
  delete_protection  = var.protected && each.value.role == "db"
  rebuild_protection = var.protected && each.value.role == "db"
  labels             = merge(local.common_labels, { "hypedomino.io/role" = each.value.role })

  user_data = "#cloud-config\n${yamlencode({
    write_files = concat([
      {
        path        = "/etc/rancher/k3s/config.yaml"
        permissions = "0600"
        content = yamlencode({
          token         = var.k3s_token
          server        = "https://${local.control_private_ip}:6443"
          "node-label"  = ["hypedomino.io/role=${each.value.role}"]
          "kubelet-arg" = local.kubelet_args
        })
      },
      { path = "/usr/local/sbin/hypedomino-install-k3s", permissions = "0700", content = local.install_script },
    ], each.value.role == "game" ? [local.game_sysctl_file] : [])
    runcmd = concat(
      each.value.role == "game" ? [["sysctl", "-p", local.game_sysctl_file.path]] : [],
      [["/usr/local/sbin/hypedomino-install-k3s", "agent"]],
    )
  })}"

  public_net {
    ipv4_enabled = true
    ipv6_enabled = true
  }

  network {
    network_id = hcloud_network.cluster.id
  }

  lifecycle {
    ignore_changes = [user_data, image, ssh_keys]
  }

  depends_on = [hcloud_network_subnet.nodes, hcloud_server.control]
}

resource "hcloud_uploaded_certificate" "game" {
  name        = "${local.name_prefix}-game-${substr(sha256(var.tls_certificate_pem), 0, 12)}"
  certificate = var.tls_certificate_pem
  private_key = var.tls_private_key_pem
  labels      = local.common_labels

  lifecycle {
    create_before_destroy = true
  }
}

resource "hcloud_load_balancer" "game" {
  name               = "${local.name_prefix}-game"
  load_balancer_type = var.load_balancer.type
  location           = var.load_balancer.location
  delete_protection  = var.protected
  labels             = local.common_labels
}

resource "hcloud_load_balancer_network" "game" {
  load_balancer_id = hcloud_load_balancer.game.id
  network_id       = hcloud_network.cluster.id

  depends_on = [hcloud_network_subnet.nodes]
}

resource "hcloud_load_balancer_target" "game" {
  for_each = var.game_nodes

  load_balancer_id = hcloud_load_balancer.game.id
  type             = "server"
  server_id        = hcloud_server.agent[each.key].id
  use_private_ip   = true

  depends_on = [hcloud_load_balancer_network.game]
}

resource "hcloud_load_balancer_service" "game_https" {
  load_balancer_id = hcloud_load_balancer.game.id
  protocol         = "https"
  listen_port      = 443
  destination_port = var.game_node_port

  http {
    certificates = [hcloud_uploaded_certificate.game.id]
  }

  health_check {
    protocol = "http"
    port     = var.game_node_port
    interval = 10
    timeout  = 5
    retries  = 3

    http {
      path         = "/readyz"
      status_codes = ["200"]
    }
  }
}

resource "hcloud_zone_rrset" "game" {
  for_each = var.dns_zone == null ? {} : {
    A    = hcloud_load_balancer.game.ipv4
    AAAA = hcloud_load_balancer.game.ipv6
  }

  zone = var.dns_zone
  name = var.game_domain == var.dns_zone ? "@" : trimsuffix(var.game_domain, ".${var.dns_zone}")
  type = each.key
  ttl  = 300

  records = [{ value = each.value }]
}
