data "google_project" "project" {
  project_id = local.gcp_context.project_id
}

data "google_compute_default_service_account" "default" {
  project = local.gcp_context.project_id
}

# To allow Cloud Run to invoke the service without authentication, AKA, Public Access
data "google_iam_policy" "noauth" {
  binding {
    role = "roles/run.invoker"
    members = [
      "allUsers",
    ]
  }
}
