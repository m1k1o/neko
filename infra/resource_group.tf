resource "azurerm_resource_group" "neko" {
  name     = var.resource_group_name != "" ? var.resource_group_name : "rg-neko-westus2"
  location = var.location
  tags     = var.tags
}

output "resource_group_id" {
  description = "ID of the resource group"
  value       = azurerm_resource_group.neko.id
}
