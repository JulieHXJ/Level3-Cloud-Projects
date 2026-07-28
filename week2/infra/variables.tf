variable "project_id" {
  description = "STACKIT project UUID in which the SKE cluster will be created"
  type        = string
}

variable "region" {
  description = "STACKIT region"
  type        = string
  default     = "eu01"
}

variable "cluster_name" {
  type    = string
  default = "tf-ske-clu"
}

variable "node_pool_name" {
  description = "Name of the SKE worker node pool"
  type        = string
  default     = "learning-pool"
}

variable "availability_zone" {
  description = "Availability zone for the single-node learning cluster"
  type        = string
  default     = "eu01-1"
}

variable "machine_type" {
  description = "Machine type used by the SKE worker node"
  type        = string
  default     = "g2i.2"
}