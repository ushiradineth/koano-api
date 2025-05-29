resource "google_project_service" "compute" {
  project                    = local.gcp_context.project_id
  service                    = "compute.googleapis.com"
  disable_on_destroy         = true
  disable_dependent_services = true
}
