
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



# output "ssh_private_key_path" {
#   value       = local_sensitive_file.ssh_private_key.filename
#   description = "auto generated SSH private key"
# }

# output "ssh_command" {
#   value = join(" ", [
#     "ssh",
#     "-i",
#     local_sensitive_file.ssh_private_key.filename,
#     "${var.ssh_user}@${openstack_networking_floatingip_v2.vm_floating_ip.address}"
#   ])
# }

# output "kubernetes_check_command" {
#   value = join(" ", [
#     "ssh",
#     "-i",
#     local_sensitive_file.ssh_private_key.filename,
#     "${var.ssh_user}@${openstack_networking_floatingip_v2.vm_floating_ip.address}",
#     "'sudo k3s kubectl get nodes -o wide'"
#   ])
# }