
# Security Group
resource "openstack_networking_secgroup_v2" "vm_security_group" {
  name        = "tf-security-group"
  description = "Security group rules for terraform-managed K3s vm"
}

resource "openstack_networking_secgroup_rule_v2" "ssh_ingress" {
  security_group_id = openstack_networking_secgroup_v2.vm_security_group.id
  direction         = "ingress"
  ethertype         = "IPv4"

  protocol         = "tcp"
  port_range_max   = 22
  port_range_min   = 22
  remote_ip_prefix = var.ssh_allowed_cidr
}

# allow ICMP for ping
resource "openstack_networking_secgroup_rule_v2" "icmp_ingress" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "icmp"
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = openstack_networking_secgroup_v2.vm_security_group.id
}


# allow Kubernetes API port 6443
resource "openstack_networking_secgroup_rule_v2" "kubernetes_api_ingress" {
  security_group_id = openstack_networking_secgroup_v2.vm_security_group.id

  direction = "ingress"
  ethertype = "IPv4"
  protocol  = "tcp"

  port_range_min = 6443
  port_range_max = 6443

  remote_ip_prefix = var.ssh_allowed_cidr
}