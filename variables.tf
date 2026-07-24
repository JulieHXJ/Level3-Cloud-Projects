
# VM creation
variable "vm_name" {
  type    = string
  default = "terraform-vm"

}

variable "image_name" {
  type    = string
  default = "Ubuntu 24.04"
}

variable "flavor_name" {
  type    = string
  default = "m1.medium"
}


# networking
variable "public_network_name" {
  description = "Existing OpenStack external network"
  type        = string
  default     = "public"
}

variable "private_network_name" {
  description = "Terraform-managed private network"
  type        = string
  default     = "tf-private-network"
}

variable "subnet_name" {
  description = "Terraform-managed subnet"
  type        = string
  default     = "tf-subnet"
}

variable "subnet_cidr" {
  description = "IPv4 CIDR of the private subnet"
  type        = string
  default     = "192.168.77.0/24"
}


# SSH login
variable "ssh_user" {
  description = "Default SSH user of the Ubuntu cloud image"
  type        = string
  default     = "ubuntu"
}

variable "ssh_allowed_cidr" {
  description = "CIDR allowed to SSH into the VM"
  type        = string

  # For now allow all ipv4
  default = "0.0.0.0/0"
}

# k8s
variable "k3s_channel" {
  description = "K3s release channel"
  type        = string
  default     = "stable"
}