terraform {
  required_version = ">= 1.13.0, < 2.0.0"

  required_providers {
    hcloud = {
      source  = "hetznercloud/hcloud"
      version = "1.70.2"
    }
  }

  backend "s3" {
    bucket                      = "hypedomino-tofu-state-production"
    key                         = "cluster/terraform.tfstate"
    region                      = "nbg1"
    endpoints                   = { s3 = "https://nbg1.your-objectstorage.com" }
    use_path_style              = true
    use_lockfile                = true
    skip_credentials_validation = true
    skip_region_validation      = true
    skip_requesting_account_id  = true
    skip_metadata_api_check     = true
    skip_s3_checksum            = true
  }

  encryption {
    key_provider "pbkdf2" "state" {
      passphrase = var.state_passphrase
    }

    method "aes_gcm" "state" {
      keys = key_provider.pbkdf2.state
    }

    state {
      method   = method.aes_gcm.state
      enforced = true
    }

    plan {
      method   = method.aes_gcm.state
      enforced = true
    }
  }
}

provider "hcloud" {}

module "cluster" {
  source = "../../modules/cluster"

  environment = "production"
  protected   = true

  k3s_token           = var.k3s_token
  ssh_public_keys     = var.ssh_public_keys
  ssh_allowed_cidrs   = var.ssh_allowed_cidrs
  api_allowed_cidrs   = var.api_allowed_cidrs
  game_domain         = var.game_domain
  game_node_port      = var.game_node_port
  tls_certificate_pem = var.tls_certificate_pem
  tls_private_key_pem = var.tls_private_key_pem
  dns_zone            = var.dns_zone
  control_node        = var.control_node
  db_node             = var.db_node
  game_nodes          = var.game_nodes
}
