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



# read existing openstack resource (Glance, Nova, Neutron)
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