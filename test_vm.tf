# terraform {
#     required_version = ">= 1.15.0, < 2.0.0"

#     required_providers {
#         openstack= {
#             source = "terraform-provider-openstack/openstack"
#             version = "3.4.0"
#         }
#     }
# }


# # variables will be read from shell 
# provider "openstack" {
# }



# read existing openstack resource as DATA (Glance, Nova, Neutron)
# data "openstack_images_image_v2" "ubuntu" {
#   name        = "Ubuntu 24.04"
#   most_recent = true
# }

# data "openstack_compute_flavor_v2" "medium" {
#   name = "m1.medium"
# }

# data "openstack_networking_network_v2" "private" {
#   name = "private"
# }

# RESOURCE is created and managed by terraform
# creat a virtual network 
resource "openstack_networking_network_v2" "private_network" {
  name           = "julie-tf-network"
  admin_state_up = true
}

# create ipv4 subnet by Neutron with DNS server
resource "openstack_networking_subnet_v2" "private_subnet" {
  name       = "julie-tf-subnet"
  network_id = openstack_networking_network_v2.private_network.id # dependency

  cidr       = "192.168.77.0/24"
  ip_version = 4

  enable_dhcp = true

  dns_nameservers = [
    "1.1.1.1",
    "8.8.8.8",
  ]

}




# Creating the vm
resource "openstack_compute_instance_v2" "vm" {
  name            = "terraform-vm-flat"
  image_name      = "Ubuntu 24.04"
  flavor_name     = "m1.medium"
  key_pair        = "level3-stackit-key"
  security_groups = ["default"]

  network {
    name = "private"
  }
}










# Output informations after terraform finishes
output "vm_name" {
  value = openstack_compute_instance_v2.vm.name
}

output "vm_status" {
  value = openstack_compute_instance_v2.vm.power_state
}

output "vm_private_ip" {
  value = openstack_compute_instance_v2.vm.access_ip_v4
}

# output "vm_network_addresses" {
#   value = openstack_compute_instance_v2.vm.all_addresses
# }