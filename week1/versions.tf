terraform {
  required_version = ">= 1.15.0, < 2.0.0"

  required_providers {
    openstack = {
      source  = "terraform-provider-openstack/openstack"
      version = "3.4.0"
    }

    # generate ssh
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.1"
    }

    # save ssh private key locally
    local = {
      source  = "hashicorp/local"
      version = "~> 2.5"
    }
  }
}

provider "openstack" {
  # Authentication information is read from OS_* environment variables.
  # may change to yaml file later

}