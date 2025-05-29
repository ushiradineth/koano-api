resource "google_artifact_registry_repository" "koano" {
  project     = local.gcp_context.project_id
  description = local.description
  location    = local.gcp_context.location

  repository_id = "koano"
  format        = "DOCKER"
}
