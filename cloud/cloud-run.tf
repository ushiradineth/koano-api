locals {
  API_VARIABLES = {
    ENV                 = var.API_ENV
    PG_USER             = var.API_PG_USER
    PG_PASSWORD         = var.API_PG_PASSWORD
    PG_URL              = var.API_PG_URL
    PG_DATABASE         = var.API_PG_DATABASE
    PG_SSLMODE          = var.API_PG_SSLMODE
    JWT_SECRET          = var.API_JWT_SECRET
    CORS_ENABLED        = var.API_CORS_ENABLED
    CORS_ALLOWED_ORIGIN = var.API_CORS_ALLOWED_ORIGIN
  }
}

resource "google_cloud_run_service" "api" {
  name     = "api"
  location = local.gcp_context.location

  template {
    spec {
      containers {
        image = "gcr.io/cloudrun/hello" # placeholder
        dynamic "env" {
          for_each = local.API_VARIABLES
          content {
            name  = env.key
            value = env.value
          }
        }
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
