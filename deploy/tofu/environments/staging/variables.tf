variable "state_passphrase" {
  type      = string
  sensitive = true

  validation {
    condition     = length(var.state_passphrase) >= 32
    error_message = "state_passphrase must be at least 32 characters."
  }
}

variable "k3s_token" {
  type      = string
  sensitive = true
}

variable "ssh_public_keys" {
  type = map(string)
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
}

variable "control_node" {
  type = object({
    server_type = string
    location    = string
  })
  default = null
}

variable "db_node" {
  type = object({
    server_type = string
    location    = string
  })
  default = null
}

variable "game_nodes" {
  type = map(object({
    server_type = string
    location    = string
  }))
  default = null
}
