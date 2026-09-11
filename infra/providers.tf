terraform {
  required_version = ">= 1.5.0"

  backend "azurerm" {
    resource_group_name  = "rg-wus2-base-infra-01"
    storage_account_name = "stgwus2baseinfra01"
    container_name       = "terraform-workload-tfstate"
    key                  = "WUS2-DEV-neko.terraform.tfstate"
  }

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }
}

provider "azurerm" {
  features {}
}
