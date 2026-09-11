output "resource_group_name" {
  description = "Name of the resource group"
  value       = azurerm_resource_group.neko.name
}

output "resource_group_location" {
  description = "Azure region of the resource group"
  value       = azurerm_resource_group.neko.location
}

output "container_app_name" {
  description = "Name of the Container Apps instance"
  value       = azurerm_container_app.neko.name
}
