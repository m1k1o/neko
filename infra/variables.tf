variable "environment_name" {
  description = "Environment name (used as prefix for resources)"
  type        = string
}

variable "location" {
  description = "Azure region for all resources"
  type        = string
}

variable "resource_group_name" {
  description = "Name of the Azure resource group"
  type        = string
}


variable "container_app_name" {
  description = "Name of the Container Apps instance running Neko"
  type        = string
}

variable "neko_image" {
  description = "Docker image to deploy (override to use your fork)"
  type        = string
}

variable "neko_http_port" {
  description = "HTTP port for Neko web interface"
  type        = number
}

variable "neko_webrtc_port_start" {
  description = "Start of UDP port range for WebRTC media"
  type        = number
}

variable "neko_webrtc_port_end" {
  description = "End of UDP port range for WebRTC media"
  type        = number
}

variable "neko_shm_size" {
  description = "Shared memory size for the browser"
  type        = string
}

variable "neko_cpu_cores" {
  description = "CPU cores allocated to the container"
  type        = number
}

variable "neko_memory_gb" {
  description = "Memory allocated to the container in GB"
  type        = number
}

variable "min_replicas" {
  description = "Minimum replicas (0 for cost savings)"
  type        = number
}

variable "max_replicas" {
  description = "Maximum replicas"
  type        = number
}

variable "enable_https" {
  description = "Enable HTTPS on the Container App ingress"
  type        = bool
}

variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
}
