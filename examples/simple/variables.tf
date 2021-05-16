variable "hostname" {
  type = string
}

variable "username" {
  type = string
  default = "root"
}

variable "port" {
  type = number
  default = null
}

variable "password" {
  type = string
  default = null
}

variable "client_private_key_pem_path" {
  type = string
  default = null
}
