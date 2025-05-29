locals {
  project = "koano"

  gcp_context = {
    project_id    = "koano-engine"
    environment   = "prod"
    location      = "asia-southeast1"
    location_abbr = "ase1"
    zone          = "asia-southeast1-a"
  }

  description = "Provisioned through Terraform by Ushira Dineth. Environment: ${local.gcp_context.environment}, Location: ${local.gcp_context.location}"
}
