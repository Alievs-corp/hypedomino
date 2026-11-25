variable "environment" {
  type = string

  validation {
    condition     = contains(["staging", "production"], var.environment)
    error_message = "environment must be staging or production."
  }
}

variable "protected" {
  type = bool
}

variable "image" {
  type     = string
  default  = "ubuntu-24.04"
  nullable = false
}

variable "control_node" {
  type = object({
    server_type = string
    location    = string
  })
  default = {
    server_type = "cx23"
    location    = "fsn1"
  }
  nullable = false
}

variable "db_node" {
  type = object({
    server_type = string
    location    = string
  })
  default = {
    server_type = "cx33"
    location    = "fsn1"
  }
  nullable = false
}

variable "game_nodes" {
  type = map(object({
    server_type = string
    location    = string
  }))
  default = {
    game-1 = {
      server_type = "cx33"
      location    = "fsn1"
    }
  }
  nullable = false

  validation {
    condition     = length(var.game_nodes) > 0
    error_message = "At least one game node is required."
  }

  validation {
    condition     = alltrue([for name in keys(var.game_nodes) : can(regex("^game-[a-z0-9-]+$", name))])
    error_message = "Game node names start with game- and use lowercase letters, digits and hyphens."
  }
}

variable "load_balancer" {
  type = object({
    type     = string
    location = string
  })
  default = {
    type     = "lb11"
    location = "fsn1"
  }
  nullable = false
}

variable "k3s_token" {
  type      = string
  sensitive = true

  validation {
    condition     = length(var.k3s_token) >= 32
    error_message = "k3s_token must be at least 32 characters."
  }
}

variable "ssh_public_keys" {
  type = map(string)

  validation {
    condition     = length(var.ssh_public_keys) > 0
    error_message = "At least one SSH public key is required."
  }
}

variable "ssh_allowed_cidrs" {
  type = list(string)
}

variable "api_allowed_cidrs" {
  type = list(string)
}

variable "game_domain" {
  type = string
}

variable "game_node_port" {
  type = number

  validation {
    condition     = var.game_node_port >= 30000 && var.game_node_port <= 32767
    error_message = "game_node_port must be in the Kubernetes NodePort range 30000-32767."
  }
}

variable "tls_certificate_pem" {
  type = string
}

variable "tls_private_key_pem" {
  type      = string
  sensitive = true
}

variable "dns_zone" {
  type    = string
  default = null

  validation {
    condition     = var.dns_zone == null || endswith(var.game_domain, var.dns_zone)
    error_message = "game_domain must be inside dns_zone."
  }
}
