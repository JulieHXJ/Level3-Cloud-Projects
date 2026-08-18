terraform {
  required_version = ">= 1.5.0"

  required_providers {
    stackit = {
      source  = "stackitcloud/stackit"
      version = "0.104.0"
    }
  }
}

provider "stackit" {
  default_region = var.region
}


# from examples
# data "stackit_ske_cluster" "example" {
#   project_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
#   name       = "example-name"
# }

resource "stackit_ske_cluster" "week2" {
  project_id = var.project_id
  name       = var.cluster_name

  node_pools = [
    {
      name                    = var.node_pool_name
      machine_type            = var.machine_type
      minimum                 = 3
      maximum                 = 5 
      availability_zones      = [var.availability_zone]
      allow_system_components = true

      cri         = "containerd" # container runtime
      os_name     = "flatcar"
      volume_size = 20
      volume_type = "storage_premium_perf1"

      max_surge       = 1
      max_unavailable = 0
    }
  ]

  network = {
    control_plane = {
      access_scope = "PUBLIC" #reachable by public ip
    }
  }
}

# a request to SKE API
resource "stackit_ske_kubeconfig" "week2" {
  project_id   = var.project_id
  cluster_name = stackit_ske_cluster.week2.name

  # 180 days: maximum supported static kubeconfig validity
  expiration = 15552000

  refresh = true

  # Refresh seven days before expiration
  refresh_before = 604800
}


output "cluster_name" {
  value = stackit_ske_cluster.week2.name
}

output "kubernetes_version_used" {
  value = stackit_ske_cluster.week2.kubernetes_version_used
}

output "node_pool_name" {
  value = var.node_pool_name
}

output "kubeconfig_expires_at" {
  value = stackit_ske_kubeconfig.week2.expires_at
}

output "kubeconfig" {
  description = "Short-lived admin kubeconfig"
  value       = stackit_ske_kubeconfig.week2.kube_config
  sensitive   = true
}