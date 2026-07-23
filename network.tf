data "openstack_networking_network_v2" "public" {
  name     = var.public_network_name
  external = true
}

# RESOURCE is created and managed by terraform
# creat a virtual network 
resource "openstack_networking_network_v2" "private_network" {
  name           = var.private_network_name
  admin_state_up = true
}

# create ipv4 subnet by Neutron with DNS server
resource "openstack_networking_subnet_v2" "private_subnet" {
  name       = var.subnet_name
  network_id = openstack_networking_network_v2.private_network.id # dependency

  cidr       = var.subnet_cidr
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