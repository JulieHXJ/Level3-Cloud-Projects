



# read existing openstack resource as DATA (Glance, Nova, Neutron)
# data "openstack_images_image_v2" "ubuntu" {
#   name        = "Ubuntu 24.04"
#   most_recent = true
# }

# data "openstack_compute_flavor_v2" "medium" {
#   name = "m1.medium"
# }

data "openstack_networking_network_v2" "public" {
  name     = "public"
  external = true
}

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

# create router and connect to private subnet
resource "openstack_networking_router_v2" "router" {
  name                = "julie-tf-router"
  admin_state_up      = true
  external_network_id = data.openstack_networking_network_v2.public.id

}

resource "openstack_networking_router_interface_v2" "private_interface" {
  router_id = openstack_networking_router_v2.router.id
  subnet_id = openstack_networking_subnet_v2.private_subnet.id
}




# Security Group
resource "openstack_networking_secgroup_v2" "vm_security_group" {
  name = "julie-tf-security-group"
  description = "rules for terraform managed vm"
}

resource "openstack_networking_secgroup_rule_v2" "ssh_ingress" {
  security_group_id = openstack_networking_secgroup_v2.vm_security_group.id
  direction = "ingress"
  ethertype = "IPv4"

  protocol = "tcp"
  port_range_max = 22
  port_range_min = 22
  remote_ip_prefix = "0.0.0.0/0"
}


#ICMP for ping



# create port
resource "openstack_networking_port_v2" "vm_port" {
  network_id = openstack_networking_network_v2.private_network.id

  name = "julie-tf-vm-port"
  admin_state_up = true

  security_group_ids = [
    openstack_networking_secgroup_v2.vm_security_group.id,
  ]

  fixed_ip {
    subnet_id = openstack_networking_subnet_v2.private_subnet.id
  }
}






# Creating the vm
resource "openstack_compute_instance_v2" "vm" {
  name            = "terraform-vm-flat"
  image_name      = "Ubuntu 24.04"
  flavor_name     = "m1.medium"
  key_pair        = "level3-stackit-key"
  # security_groups = ["default"]

  network {
    # name = "private"
    port = openstack_networking_port_v2.vm_port.id
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