variable "project_name" {
  description = "Name of the project"
  type        = string
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
}

variable "artifact_retention_days" {
  description = "Number of days to retain Jenkins artifacts"
  type        = number
  default     = 30
}

variable "backup_retention_days" {
  description = "Number of days to retain database backups"
  type        = number
  default     = 90
}

variable "log_retention_days" {
  description = "Number of days to retain application logs"
  type        = number
  default     = 30
}