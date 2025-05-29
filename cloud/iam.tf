resource "google_service_account" "github_action" {
  project     = local.gcp_context.project_id
  description = local.description

  account_id   = "github-action"
  display_name = "GitHub Actions Cloud Run Deployer"
}

resource "google_project_iam_member" "github_actions_cloudrun_admin" {
  project = local.gcp_context.project_id

  role   = "roles/run.admin"
  member = "serviceAccount:${google_service_account.github_action.email}"

  depends_on = [google_service_account.github_action]
}

resource "google_project_iam_member" "github_actions_iam_user" {
  project = local.gcp_context.project_id

  role   = "roles/iam.serviceAccountUser"
  member = "serviceAccount:${google_service_account.github_action.email}"

  depends_on = [google_service_account.github_action]
}

resource "google_artifact_registry_repository_iam_binding" "github_action_binding" {
  project    = local.gcp_context.project_id
  location   = local.gcp_context.location
  repository = google_artifact_registry_repository.koano.name
  role       = "roles/artifactregistry.repoAdmin"
  members    = ["serviceAccount:${google_service_account.github_action.email}"]

  depends_on = [google_service_account.github_action, google_artifact_registry_repository.koano]
}
