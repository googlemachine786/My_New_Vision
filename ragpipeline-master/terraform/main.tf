# Visionary RAG Pipeline - GCP Infrastructure
# Phase 1 - Production Infrastructure for CBSE Science RAG System
# Region: asia-south1 | AlloyDB REGIONAL | Cloud Run min=2

terraform {
  required_version = ">= 1.6.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }

  backend "gcs" {
    # Bucket for Terraform state - create manually before first init
    # bucket = "visionary-rag-terraform-state"
    prefix = "terraform/state"
  }
}

provider "google" {
  project = var.gcp_project_id
  region  = var.region
}

# =============================================================================
# VPC Network
# =============================================================================

resource "google_compute_network" "vpc" {
  name                            = "visionary-rag-vpc"
  project                         = var.gcp_project_id
  auto_create_subnetworks         = false
  routing_mode                    = "REGIONAL"
  delete_default_routes_on_create = false
}

# =============================================================================
# Subnet for Cloud Run VPC Connector
# =============================================================================

resource "google_compute_subnetwork" "cloud_run" {
  name          = "visionary-rag-subnet"
  project       = var.gcp_project_id
  region        = var.region
  network       = google_compute_network.vpc.id
  ip_cidr_range = "10.8.0.0/24"
}

# =============================================================================
# VPC Access Connector - Bridge for Cloud Run to VPC
# =============================================================================

resource "google_vpc_access_connector" "connector" {
  name          = "visionary-rag-vpc-connector"
  project       = var.gcp_project_id
  region        = var.region
  subnet {
    name = google_compute_subnetwork.cloud_run.name
  }
  machine_type = "e2-standard"
  min_instances = 2
  max_instances = 10
}

# =============================================================================
# AlloyDB Cluster - REGIONAL for 99.99% SLA
# =============================================================================

resource "google_alloydb_cluster" "visionary" {
  cluster_id = "visionary-rag-cluster"
  location   = var.region
  project    = var.gcp_project_id

  network_config {
    network    = google_compute_network.vpc.id
    allocated_ip_range = "visionary-rag-subnet"
  }

  initial_user {
    user     = "visionary"
    password = var.alloydb_initial_password
  }

  automated_backup_policy {
    location      = "REGIONAL"
    backup_window = "04:00:00"
    enabled       = true

    weekly_schedule {
      preferred_days {
        day_of_week = "MONDAY"
      }
      preferred_days {
        day_of_week = "WEDNESDAY"
      }
      preferred_days {
        day_of_week = "FRIDAY"
      }
    }
  }

  labels = {
    environment = "production"
    application = "rag-pipeline"
  }
}

# =============================================================================
# AlloyDB Primary Instance - 8 vCPU
# =============================================================================

resource "google_alloydb_instance" "primary" {
  instance_id = "visionary-rag-primary"
  cluster     = google_alloydb_cluster.visionary.name
  location    = var.region
  project     = var.gcp_project_id

  machine_config {
    machine_type = "n2-standard-8"  # 8 vCPU, 32GB RAM
  }

  database_flags = {
    # Enable pgvector extension support
    "pgvector.enable" = "on"
    # Max connections tuned for PgBouncer pooling
    "max_connections" = "20"
  }

  labels = {
    environment = "production"
    role        = "primary"
  }
}

# =============================================================================
# Memorystore Redis 7.x - Session Store
# =============================================================================

resource "google_redis_instance" "session" {
  name                           = "visionary-rag-session"
  project                        = var.gcp_project_id
  region                         = var.region
  tier                           = "STANDARD_HA"
  memory_size_gb                 = 4
  redis_version                  = "REDIS_7_0"
  display_name                   = "Visionary RAG Session Store"
  authorized_network             = google_compute_network.vpc.id
  connect_mode                   = "DIRECT_PEERING"
  
  # Critical: 60% of RAM cap prevents OOM kill
  redis_configs = {
    "maxmemory-policy" = "allkeys-lru"
    "maxmemory"        = "2621440kb"  # 2.5GB (60% of 4GB)
  }

  labels = {
    environment = "production"
    application = "rag-pipeline"
  }

  # Maintenance window during low traffic
  maintenance_policy {
    weekly_maintenance_window {
      day = "SUNDAY"
      start_time {
        hours   = 3
        minutes = 0
        seconds = 0
        nanos   = 0
      }
    }
  }
}

# =============================================================================
# Cloud Run Service - Go Orchestrator
# =============================================================================

resource "google_cloud_run_v2_service" "orchestrator" {
  name     = "visionary-rag-orchestrator"
  location = var.region
  project  = var.gcp_project_id

  ingress = "INGRESS_TRAFFIC_INTERNAL_LOAD_BALANCER_ONLY"

  template {
    # CRITICAL: Never scale to zero for TTFT SLA
    min_instance_count = 2
    max_instance_count = 100
    
    scaling {
      target_cpu_utilization           = 70
      target_memory_utilization        = 80
      max_instance_request_concurrency = 100
    }

    # CRITICAL: Keep CPU active during idle for warm starts
    scaling_policy {
      target_cpu_utilization = 70
    }

    volumes {
      name = "pgbouncer-socket"
      empty_dir {
        medium = "MEMORY"
      }
    }

    containers {
      image = "${var.region}-docker.pkg.dev/${var.gcp_project_id}/${google_artifact_registry_repository.images.repository_id}/orchestrator:latest"
      
      ports {
        name           = "http"
        container_port = 8080
      }

      env {
        name  = "PORT"
        value = "8080"
      }

      env {
        name = "ALLOYDB_DSN"
        value_source {
          secret_key_ref {
            secret  = "alloydb-dsn"
            version = "latest"
          }
        }
      }

      env {
        name = "REDIS_ADDR"
        value_source {
          secret_key_ref {
            secret  = "redis-addr"
            version = "latest"
          }
        }
      }

      env {
        name = "VERTEX_PROJECT"
        value = var.gcp_project_id
      }

      env {
        name = "VERTEX_LOCATION"
        value = var.region
      }

      env {
        name = "JWT_SECRET"
        value_source {
          secret_key_ref {
            secret  = "jwt-secret"
            version = "latest"
          }
        }
      }

      resources {
        cpu    = 2
        memory = "2Gi"
        # GPU not needed for embedding + generation (Vertex AI handles)
      }

      # Health check
      liveness_probe {
        http_get {
          path = "/health"
          port = 8080
        }
        initial_delay_seconds = 5
        timeout_seconds      = 5
        period_seconds       = 10
        failure_threshold    = 3
      }

      readiness_probe {
        http_get {
          path = "/health"
          port = 8080
        }
        initial_delay_seconds = 3
        timeout_seconds      = 3
        period_seconds       = 5
        failure_threshold    = 2
      }
    }

    # VPC Access for AlloyDB and Redis
    vpc_access {
      connector = google_vpc_access_connector.connector.id
      egress    = "PRIVATE_RANGES_ONLY"
    }
  }

  traffic {
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
    percent = 100
  }

  labels = {
    environment = "production"
    application = "rag-pipeline"
  }
}

# =============================================================================
# Cloud Run Job - Python Ingestion Pipeline
# =============================================================================

resource "google_cloud_run_v2_job" "ingestion" {
  name     = "visionary-rag-ingestion"
  location = var.region
  project  = var.gcp_project_id

  template {
    template {
      containers {
        image = "${var.region}-docker.pkg.dev/${var.gcp_project_id}/${google_artifact_registry_repository.images.repository_id}/ingestion:latest"
        
        env {
          name = "ALLOYDB_DSN"
          value_source {
            secret_key_ref {
              secret  = "alloydb-dsn"
              version = "latest"
            }
          }
        }

        env {
          name = "VERTEX_PROJECT"
          value = var.gcp_project_id
        }

        env {
          name = "VERTEX_LOCATION"
          value = var.region
        }

        resources {
          cpu    = 2
          memory = "4Gi"
        }
      }

      vpc_access {
        connector = google_vpc_access_connector.connector.id
        egress    = "PRIVATE_RANGES_ONLY"
      }
    }
  }

  labels = {
    environment = "production"
    application = "rag-pipeline"
  }
}

# =============================================================================
# Secret Manager - 4 Required Secrets
# =============================================================================

resource "google_secret_manager_secret" "alloydb_dsn" {
  secret_id = "alloydb-dsn"
  project   = var.gcp_project_id

  replication {
    auto {}
  }

  labels = {
    environment = "production"
  }
}

resource "google_secret_manager_secret" "redis_addr" {
  secret_id = "redis-addr"
  project   = var.gcp_project_id

  replication {
    auto {}
  }

  labels = {
    environment = "production"
  }
}

resource "google_secret_manager_secret" "vertex_sa" {
  secret_id = "vertex-sa"
  project   = var.gcp_project_id

  replication {
    auto {}
  }

  labels = {
    environment = "production"
  }
}

resource "google_secret_manager_secret" "jwt_secret" {
  secret_id = "jwt-secret"
  project   = var.gcp_project_id

  replication {
    auto {}
  }

  labels = {
    environment = "production"
  }
}

# =============================================================================
# Artifact Registry - Container Images
# =============================================================================

resource "google_artifact_registry_repository" "images" {
  location      = var.region
  repository_id = "visionary-rag-images"
  description   = "Container images for Visionary RAG pipeline"
  format        = "DOCKER"

  cleanup_policy_dry_run = false
}

# =============================================================================
# IAM Bindings - Service Account Permissions
# =============================================================================

# Cloud Run service account needs access to Secret Manager
resource "google_project_iam_member" "orchestrator_secrets" {
  project = var.gcp_project_id
  role    = "roles/secretmanager.secretAccessor"
  member  = "serviceAccount:${var.gcp_project_id}.svc.id.goog[visionary-rag-orchestrator]"
}

# Cloud Run service account needs Vertex AI access
resource "google_project_iam_member" "orchestrator_vertexai" {
  project = var.gcp_project_id
  role    = "roles/aiplatform.user"
  member  = "serviceAccount:${var.gcp_project_id}.svc.id.goog[visionary-rag-orchestrator]"
}

# Cloud Run service account needs Cloud Trace access
resource "google_project_iam_member" "orchestrator_trace" {
  project = var.gcp_project_id
  role    = "roles/cloudtrace.agent"
  member  = "serviceAccount:${var.gcp_project_id}.svc.id.goog[visionary-rag-orchestrator]"
}
