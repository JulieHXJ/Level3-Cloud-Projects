# # create key pair
# resource "tls_private_key" "vm_ssh_key" {
#   algorithm = "ED25519"
# }

# resource "openstack_compute_keypair_v2" "vm_keypair" {
#   name       = "${var.vm_name}-key"
#   public_key = tls_private_key.vm_ssh_key.public_key_openssh
# }

# # save private key to server
# resource "local_sensitive_file" "ssh_private_key" {
#   filename        = "${path.module}/generated/${var.vm_name}.pem"
#   content         = tls_private_key.vm_ssh_key.private_key_openssh
#   file_permission = "0600"
# }



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


# Floating ip
resource "openstack_networking_floatingip_v2" "vm_floating_ip" {
  pool    = data.openstack_networking_network_v2.public.name
  port_id = openstack_networking_port_v2.vm_port.id

  depends_on = [openstack_networking_router_interface_v2.private_interface] #setup router connection with fixed ip
}






# Creating the vm
resource "openstack_compute_instance_v2" "vm" {
  name        = var.vm_name
  image_name  = var.image_name
  flavor_name = var.flavor_name
  # key_pair    = openstack_compute_keypair_v2.k3s_keypair.name
  key_pair = "level3_stackit_key"

  # # use cloud config yaml and do cloud init
  # user_data = templatefile(
  #   "${path.module}/cloud_config.yaml.tf"

  # )

  network {
    port = openstack_networking_port_v2.vm_port.id
  }

  depends_on = [
    openstack_networking_secgroup_rule_v2.ssh_ingress,
    openstack_networking_secgroup_rule_v2.icmp_ingress
  ]
}



