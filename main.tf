terraform {
  required_version = ">= 1.15.0, < 2.0.0"

  required_providers {
    openstack = {
      source  = "terraform-provider-openstack/openstack"
      version = "3.4.0"
    }
  }
}


# env variables will be read from shell 
provider "openstack" {
}



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
  name           = "tf-private-network"
  admin_state_up = true
}

# create ipv4 subnet by Neutron with DNS server
resource "openstack_networking_subnet_v2" "private_subnet" {
  name       = "tf-subnet"
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
  name                = "tf-router"
  admin_state_up      = true
  external_network_id = data.openstack_networking_network_v2.public.id

}

resource "openstack_networking_router_interface_v2" "private_interface" {
  router_id = openstack_networking_router_v2.router.id
  subnet_id = openstack_networking_subnet_v2.private_subnet.id
}




# Security Group
resource "openstack_networking_secgroup_v2" "vm_security_group" {
  name        = "tf-security-group"
  description = "rules for terraform managed vm"
}

resource "openstack_networking_secgroup_rule_v2" "ssh_ingress" {
  security_group_id = openstack_networking_secgroup_v2.vm_security_group.id
  direction         = "ingress"
  ethertype         = "IPv4"

  protocol         = "tcp"
  port_range_max   = 22
  port_range_min   = 22
  remote_ip_prefix = "0.0.0.0/0" #allow all ipv4
}





# create port
resource "openstack_networking_port_v2" "vm_port" {
  network_id = openstack_networking_network_v2.private_network.id

  name           = "tf-vm-port"
  admin_state_up = true

  security_group_ids = [
    openstack_networking_secgroup_v2.vm_security_group.id,
  ]

  fixed_ip {
    subnet_id = openstack_networking_subnet_v2.private_subnet.id
  }
}


# allow ICMP for ping
resource "openstack_networking_secgroup_rule_v2" "icmp_ingress" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "icmp"
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = openstack_networking_secgroup_v2.vm_security_group.id
}

# Floating ip
resource "openstack_networking_floatingip_v2" "vm_floating_ip" {
  pool    = data.openstack_networking_network_v2.public.name
  port_id = openstack_networking_port_v2.vm_port.id

  depends_on = [openstack_networking_router_interface_v2.private_interface] #setup router connection with fixed ip
}



# Creating the vm
resource "openstack_compute_instance_v2" "vm" {
  name        = "terraform-vm-flat"
  image_name  = "Ubuntu 24.04"
  flavor_name = "m1.medium"
  key_pair    = "level3-stackit-key" #todo: auto generate ssh key pair on each run


  network {

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
  value = openstack_networking_port_v2.vm_port.all_fixed_ips[0]
}

output "vm_floating_ip" {
  value       = openstack_networking_floatingip_v2.vm_floating_ip.address
  description = "Floating IP for VM access"
}

