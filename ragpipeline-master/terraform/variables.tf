# Visionary RAG Pipeline - Terraform Variables
# Phase 1 - Production Infrastructure

# =============================================================================
# GCP Project Configuration
# =============================================================================

variable "gcp_project_id" {
  description = "GCP project ID for the Visionary RAG pipeline"
  type        = string
  nullable    = false

  validation {
    condition     = can(regex("^[a-z][a-z0-9-]{4,28}[a-z0-9]$", var.gcp_project_id))
    error_message = "Project ID must be 6-30 characters, start with letter, contain only lowercase letters, numbers, and hyphens."
  }
}

# =============================================================================
# Region Configuration
# =============================================================================

variable "region" {
  description = "GCP region for resource deployment (asia-south1 recommended for India)"
  type        = string
  default     = "asia-south1"

  validation {
    condition     = can(regex("^[a-z]+-[a-z]+[0-9]*$", var.region))
    error_message = "Region must be a valid GCP region (e.g., asia-south1)"
  }
}

variable "zone" {
  description = "GCP zone for zonal resources (optional - REGIONAL used where available)"
  type        = string
  default     = "asia-south1-a"

  validation {
    condition     = can(regex("^[a-z]+-[a-z]+[0-9]*-[a-z]$", var.zone))
    error_message = "Zone must be a valid GCP zone (e.g., asia-south1-a)"
  }
}

# =============================================================================
# AlloyDB Configuration
# =============================================================================

variable "alloydb_initial_password" {
  description = "Initial password for AlloyDB 'visionary' user (must meet complexity requirements)"
  type        = string
  sensitive   = true
  nullable    = false

  validation {
    condition     = length(var.alloydb_initial_password) >= 8
    error_message = "Password must be at least 8 characters long."
  }
}

variable "alloydb_machine_type" {
  description = "AlloyDB instance machine type"
  type        = string
  default     = "n2-standard-8"  # 8 vCPU, 32GB RAM

  validation {
    condition     = contains(["n2-standard-2", "n2-standard-4", "n2-standard-8", "n2-standard-16", "n2-standard-32"], var.alloydb_machine_type)
    error_message = "Machine type must be a valid n2-standard size."
  }
}

# =============================================================================
# Redis Configuration
# =============================================================================

variable "redis_memory_gb" {
  description = "Redis instance memory size in GB"
  type        = number
  default     = 4

  validation {
    condition     = var.redis_memory_gb >= 1 && var.redis_memory_gb <= 50
    error_message = "Redis memory must be between 1 and 50 GB."
  }
}

variable "redis_maxmemory_policy" {
  description = "Redis maxmemory eviction policy"
  type        = string
  default     = "allkeys-lru"

  validation {
    condition     = contains(["allkeys-lru", "volatile-lru", "allkeys-random", "volatile-random", "volatile-ttl", "noeviction"], var.redis_maxmemory_policy)
    error_message = "Invalid maxmemory policy."
  }
}

# =============================================================================
# Cloud Run Configuration
# =============================================================================

variable "cloud_run_min_instances" {
  description = "Minimum number of Cloud Run instances (2 recommended for TTFT SLA)"
  type        = number
  default     = 2

  validation {
    condition     = var.cloud_run_min_instances >= 0 && var.cloud_run_min_instances <= 10
    error_message = "Min instances must be between 0 and 10."
  }
}

variable "cloud_run_max_instances" {
  description = "Maximum number of Cloud Run instances"
  type        = number
  default     = 100

  validation {
    condition     = var.cloud_run_max_instances >= 1 && var.cloud_run_max_instances <= 1000
    error_message = "Max instances must be between 1 and 1000."
  }
}

variable "cloud_run_cpu" {
  description = "Cloud Run CPU allocation"
  type        = number
  default     = 2

  validation {
    condition     = contains([1, 2, 4], var.cloud_run_cpu)
    error_message = "CPU must be 1, 2, or 4."
  }
}

variable "cloud_run_memory" {
  description = "Cloud Run memory allocation"
  type        = string
  default     = "2Gi"

  validation {
    condition     = contains(["512Mi", "1Gi", "2Gi", "4Gi", "8Gi"], var.cloud_run_memory)
    error_message = "Memory must be 512Mi, 1Gi, 2Gi, 4Gi, or 8Gi."
  }
}

# =============================================================================
# VPC Configuration
# =============================================================================

variable "vpc_cidr" {
  description = "VPC subnet CIDR for Cloud Run connector"
  type        = string
  default     = "10.8.0.0/24"

  validation {
    condition     = can(cidrhost(var.vpc_cidr, 0))
    error_message = "VPC CIDR must be a valid IPv4 CIDR block."
  }
}

# =============================================================================
# Tags and Labels
# =============================================================================

variable "environment" {
  description = "Environment label (production, staging, development)"
  type        = string
  default     = "production"

  validation {
    condition     = contains(["production", "staging", "development"], var.environment)
    error_message = "Environment must be production, staging, or development."
  }
}

variable "application" {
  description = "Application name label"
  type        = string
  default     = "rag-pipeline"
}

# =============================================================================
# Backup Configuration
# =============================================================================

variable "backup_window" {
  description = "Automated backup window (UTC time)"
  type        = string
  default     = "04:00:00"

  validation {
    condition     = can(regex("^([01]?[0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9]$", var.backup_window))
    error_message = "Backup window must be in HH:MM:SS format."
  }
}

variable "backup_retention_days" {
  description = "Number of days to retain automated backups"
  type        = number
  default     = 7

  validation {
    condition     = var.backup_retention_days >= 1 && var.backup_retention_days <= 365
    error_message = "Backup retention must be between 1 and 365 days."
  }
}
