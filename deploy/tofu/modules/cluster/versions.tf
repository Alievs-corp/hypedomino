terraform {
  required_version = ">= 1.13.0, < 2.0.0"

  required_providers {
    hcloud = {
      source  = "hetznercloud/hcloud"
      version = ">= 1.70.2, < 2.0.0"
    }
  }
}
