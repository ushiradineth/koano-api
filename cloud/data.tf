data "google_project" "project" {
  project_id = local.gcp_context.project_id
}

data "google_compute_default_service_account" "default" {
  project = local.gcp_context.project_id
}
