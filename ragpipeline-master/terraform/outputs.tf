# Visionary RAG Pipeline - Terraform Outputs
# Phase 1 - Production Infrastructure

# =============================================================================
# AlloyDB Outputs
# =============================================================================

output "alloydb_cluster_name" {
  description = "Name of the AlloyDB cluster"
  value       = google_alloydb_cluster.visionary.cluster_id
}

output "alloydb_cluster_uri" {
  description = "AlloyDB cluster resource URI"
  value       = google_alloydb_cluster.visionary.name
}

output "alloydb_primary_ip" {
  description = "AlloyDB primary instance private IP address"
  value       = google_alloydb_instance.primary.ip_address
  sensitive   = false
}

output "alloydb_connection_string" {
  description = "AlloyDB connection string (use PgBouncer in production)"
  value       = "host=${google_alloydb_instance.primary.ip_address} port=5432 dbname=visionary user=visionary password=${var.alloydb_initial_password}"
  sensitive   = true
}

output "alloydb_pgbouncer_connection_string" {
  description = "AlloyDB connection string via PgBouncer (Cloud Run sidecar)"
  value       = "host=localhost port=6432 dbname=visionary user=visionary password=${var.alloydb_initial_password}"
  sensitive   = true
}

# =============================================================================
# Redis Outputs
# =============================================================================

output "redis_host" {
  description = "Redis instance private IP address"
  value       = google_redis_instance.session.host
}

output "redis_port" {
  description = "Redis instance port"
  value       = google_redis_instance.session.port
}

output "redis_address" {
  description = "Redis connection address (host:port)"
  value       = "${google_redis_instance.session.host}:${google_redis_instance.session.port}"
}

output "redis_uri" {
  description = "Redis instance full URI"
  value       = google_redis_instance.session.id
}

# =============================================================================
# Cloud Run Outputs
# =============================================================================

output "orchestrator_service_name" {
  description = "Cloud Run orchestrator service name"
  value       = google_cloud_run_v2_service.orchestrator.name
}

output "orchestrator_service_url" {
  description = "Cloud Run orchestrator service URL"
  value       = google_cloud_run_v2_service.orchestrator.uri
}

output "orchestrator_service_location" {
  description = "Cloud Run orchestrator service location"
  value       = google_cloud_run_v2_service.orchestrator.location
}

output "ingestion_job_name" {
  description = "Cloud Run ingestion job name"
  value       = google_cloud_run_v2_job.ingestion.name
}

output "ingestion_job_location" {
  description = "Cloud Run ingestion job location"
  value       = google_cloud_run_v2_job.ingestion.location
}

# =============================================================================
# VPC Outputs
# =============================================================================

output "vpc_name" {
  description = "VPC network name"
  value       = google_compute_network.vpc.name
}

output "vpc_self_link" {
  description = "VPC network self link"
  value       = google_compute_network.vpc.self_link
}

output "subnet_name" {
  description = "Subnet name for Cloud Run connector"
  value       = google_compute_subnetwork.cloud_run.name
}

output "subnet_cidr" {
  description = "Subnet CIDR range"
  value       = google_compute_subnetwork.cloud_run.ip_cidr_range
}

output "vpc_connector_name" {
  description = "VPC Access Connector name"
  value       = google_vpc_access_connector.connector.name
}

output "vpc_connector_id" {
  description = "VPC Access Connector ID"
  value       = google_vpc_access_connector.connector.id
}

# =============================================================================
# Secret Manager Outputs
# =============================================================================

output "secret_names" {
  description = "List of created Secret Manager secret names"
  value = [
    google_secret_manager_secret.alloydb_dsn.secret_id,
    google_secret_manager_secret.redis_addr.secret_id,
    google_secret_manager_secret.vertex_sa.secret_id,
    google_secret_manager_secret.jwt_secret.secret_id
  ]
}

output "alloydb_dsn_secret_name" {
  description = "Secret Manager secret name for AlloyDB DSN"
  value       = google_secret_manager_secret.alloydb_dsn.secret_id
}

output "redis_addr_secret_name" {
  description = "Secret Manager secret name for Redis address"
  value       = google_secret_manager_secret.redis_addr.secret_id
}

output "vertex_sa_secret_name" {
  description = "Secret Manager secret name for Vertex AI service account"
  value       = google_secret_manager_secret.vertex_sa.secret_id
}

output "jwt_secret_secret_name" {
  description = "Secret Manager secret name for JWT secret"
  value       = google_secret_manager_secret.jwt_secret.secret_id
}

# =============================================================================
# Artifact Registry Outputs
# =============================================================================

output "artifact_registry_name" {
  description = "Artifact Registry repository name"
  value       = google_artifact_registry_repository.images.name
}

output "artifact_registry_url" {
  description = "Artifact Registry Docker push URL"
  value       = "${var.region}-docker.pkg.dev/${var.gcp_project_id}/${google_artifact_registry_repository.images.repository_id}"
}

# =============================================================================
# Project Configuration Outputs
# =============================================================================

output "project_id" {
  description = "GCP project ID"
  value       = var.gcp_project_id
}

output "region" {
  description = "GCP region"
  value       = var.region
}

output "zone" {
  description = "GCP zone"
  value       = var.zone
}

# =============================================================================
# Deployment Instructions
# =============================================================================

output "deployment_instructions" {
  description = "Next steps after Terraform apply"
  value       = <<-EOT
    Visionary RAG Infrastructure Deployed!
    
    Next Steps:
    1. Apply database schema:
       psql "${google_alloydb_instance.primary.ip_address}" -f schema/v2_production.sql
    
    2. Populate Secret Manager secrets:
       - alloydb-dsn: Database connection string
       - redis-addr: ${google_redis_instance.session.host}:${google_redis_instance.session.port}
       - vertex-sa: Service account key JSON
       - jwt-secret: JWT signing key
    
    3. Deploy PgBouncer sidecar to Cloud Run:
       Copy orchestrator/pgbouncer/pgbouncer.ini to Cloud Run volume
    
    4. Build and push container images:
       docker build -t ${var.region}-docker.pkg.dev/${var.gcp_project_id}/visionary-rag-images/orchestrator:latest ./orchestrator
       docker push ${var.region}-docker.pkg.dev/${var.gcp_project_id}/visionary-rag-images/orchestrator:latest
    
    5. Verify Redis latency:
       redis-cli -h ${google_redis_instance.session.host} --latency-history
    
    6. Verify database schema:
       bash scripts/verify-schema.sh
    
    Cloud Run Service URL:
    ${google_cloud_run_v2_service.orchestrator.uri}
  EOT
}
