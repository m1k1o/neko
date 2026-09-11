# Log Analytics Workspace (required by Container Apps)
resource "azurerm_log_analytics_workspace" "neko" {
  name                = "${var.environment_name}-logs"
  resource_group_name = azurerm_resource_group.neko.name
  location            = azurerm_resource_group.neko.location
  sku                 = "PerGB2018"
  retention_in_days   = 30

  tags = var.tags
}

# Container Apps Environment
resource "azurerm_container_app_environment" "neko" {
  name                       = "${var.environment_name}-env"
  resource_group_name        = azurerm_resource_group.neko.name
  location                   = azurerm_resource_group.neko.location
  logs_destination           = "log-analytics"
  log_analytics_workspace_id = azurerm_log_analytics_workspace.neko.id

  tags = var.tags
}

# Container App running Neko
resource "azurerm_container_app" "neko" {
  name                         = var.container_app_name != "" ? var.container_app_name : "${var.environment_name}-neko"
  resource_group_name          = azurerm_resource_group.neko.name
  container_app_environment_id = azurerm_container_app_environment.neko.id
  revision_mode                = "Single"

  template {
    max_replicas = var.max_replicas
    min_replicas = var.min_replicas

    container {
      name   = "neko"
      image  = var.neko_image != "" ? var.neko_image : "ghcr.io/networkneil/neko:latest"
      cpu    = tostring(var.neko_cpu_cores)
      memory = "${var.neko_memory_gb}Gi"

      # Neko environment variables
      env {
        name  = "NEKO_SERVER_BIND"
        value = ":8080"
      }
      env {
        name  = "NEKO_WEBRTC_EPR"
        value = "${var.neko_webrtc_port_start}-${var.neko_webrtc_port_end}"
      }
      env {
        name  = "NEKO_SERVER_PROXY"
        value = "true"
      }
      env {
        name  = "NO_COLOR"
        value = "1"
      }
    }
  }

  lifecycle {
    ignore_changes = [template[0].container[0].image]
  }

  tags = var.tags
}
