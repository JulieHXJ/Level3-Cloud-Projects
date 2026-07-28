# variables can be modified inside the file 
vm_name     = "terraform-k3s-vm"
image_name  = "Ubuntu 24.04"
flavor_name = "m1.medium"

private_network_name = "tf-private-network"
subnet_name          = "tf-subnet"
subnet_cidr          = "192.168.77.0/24"

ssh_user         = "ubuntu"
ssh_allowed_cidr = "0.0.0.0/0"

k3s_channel = "stable"