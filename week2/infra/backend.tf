terraform {
  backend "s3" {
    bucket = "julie-week2-terraform-state"
    key    = "week2/infra/terraform.tfstate"
    region = "eu01"

    endpoints = {
      s3 = "https://object.storage.eu01.onstackit.cloud"
    }


    skip_credentials_validation = true
    skip_region_validation      = true
    skip_s3_checksum            = true
    skip_requesting_account_id  = true
  }
}