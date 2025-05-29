terraform {
  required_version = ">= v1.12.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "6.37.0"
    }
  }

  cloud {
    organization = "shu"

    workspaces {
      name = "prod"
    }
  }
}

provider "google" {
  project = local.gcp_context.project_id
  region  = local.gcp_context.location
}
