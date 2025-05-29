resource "google_cloud_run_service" "api" {
  name     = "api"
  location = local.gcp_context.location

  template {
    spec {
      containers {
        image = "gcr.io/cloudrun/hello" # placeholder
      }
      service_account_name = google_service_account.github_action.email
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }

  lifecycle {
    ignore_changes = [
      template[0].spec[0].containers[0].image, # This will be updated by the GitHub Action
    ]
  }

  depends_on = [google_service_account.github_action]
}
