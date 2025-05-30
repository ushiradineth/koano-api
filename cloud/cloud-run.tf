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
  project  = local.gcp_context.project_id
  location = local.gcp_context.location

  template {
    spec {
      containers {
        image = "gcr.io/cloudrun/hello" # Placeholder, This will be updated by the GitHub Action
        dynamic "env" {
          for_each = local.API_VARIABLES
          content {
            name  = env.key
            value = env.value
          }
        }
        startup_probe {
          initial_delay_seconds = 0
          timeout_seconds       = 1
          period_seconds        = 3
          failure_threshold     = 1
          tcp_socket {
            port = 8080
          }
        }
        liveness_probe {
          http_get {
            path = "/health"
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

# To allow Cloud Run to invoke the service without authentication, AKA, Public Access
resource "google_cloud_run_service_iam_policy" "api" {
  project  = google_cloud_run_service.api.project
  location = google_cloud_run_service.api.location
  service  = google_cloud_run_service.api.name

  policy_data = data.google_iam_policy.noauth.policy_data

  depends_on = [google_cloud_run_service.api, data.google_iam_policy.noauth]
}
