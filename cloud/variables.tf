variable "API_ENV" {
  description = "API Environment"
  type        = string
  default     = "PRODUCTION"
}

variable "API_PG_USER" {
  description = "Postgres User"
  type        = string
  sensitive   = true
}

variable "API_PG_PASSWORD" {
  description = "Postgres Password"
  type        = string
  sensitive   = true
}

variable "API_PG_URL" {
  description = "Postgres URL"
  type        = string
  sensitive   = true
}

variable "API_PG_DATABASE" {
  description = "Postgres Database"
  type        = string
  sensitive   = true
}

variable "API_PG_SSLMODE" {
  description = "Postgres SSL Mode"
  type        = string
  sensitive   = true
}

variable "API_JWT_SECRET" {
  description = "JWT Secret"
  type        = string
  sensitive   = true
}

variable "API_CORS_ENABLED" {
  description = "CORS Enabled"
  type        = string
  sensitive   = true
}

variable "API_CORS_ALLOWED_ORIGIN" {
  description = "CORS Allowed Origin"
  type        = string
  sensitive   = true
}
