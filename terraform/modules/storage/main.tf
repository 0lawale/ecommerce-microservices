# Storage Module - S3 Buckets for artifacts, backups, and Terraform state

# S3 Bucket for Jenkins Artifacts
resource "aws_s3_bucket" "jenkins_artifacts" {
  bucket = "${var.project_name}-${var.environment}-jenkins-artifacts"

  tags = {
    Name        = "${var.project_name}-${var.environment}-jenkins-artifacts"
    Environment = var.environment
    Purpose     = "Jenkins build artifacts"
  }
}

# Enable versioning for Jenkins artifacts
resource "aws_s3_bucket_versioning" "jenkins_artifacts" {
  bucket = aws_s3_bucket.jenkins_artifacts.id

  versioning_configuration {
    status = "Enabled"
  }
}

# Server-side encryption for Jenkins artifacts
resource "aws_s3_bucket_server_side_encryption_configuration" "jenkins_artifacts" {
  bucket = aws_s3_bucket.jenkins_artifacts.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# Block public access for Jenkins artifacts
resource "aws_s3_bucket_public_access_block" "jenkins_artifacts" {
  bucket = aws_s3_bucket.jenkins_artifacts.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Lifecycle policy for Jenkins artifacts (delete old artifacts after 30 days)
resource "aws_s3_bucket_lifecycle_configuration" "jenkins_artifacts" {
  bucket = aws_s3_bucket.jenkins_artifacts.id

  rule {
    id     = "delete-old-artifacts"
    status = "Enabled"
    filter {
      prefix = ""
    }

    expiration {
      days = var.artifact_retention_days
    }

    noncurrent_version_expiration {
      noncurrent_days = 7
    }
  }
}

# S3 Bucket for Database Backups
resource "aws_s3_bucket" "database_backups" {
  bucket = "${var.project_name}-${var.environment}-db-backups"

  tags = {
    Name        = "${var.project_name}-${var.environment}-db-backups"
    Environment = var.environment
    Purpose     = "Database backups"
  }
}

# Enable versioning for database backups
resource "aws_s3_bucket_versioning" "database_backups" {
  bucket = aws_s3_bucket.database_backups.id

  versioning_configuration {
    status = "Enabled"
  }
}

# Server-side encryption for database backups
resource "aws_s3_bucket_server_side_encryption_configuration" "database_backups" {
  bucket = aws_s3_bucket.database_backups.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# Block public access for database backups
resource "aws_s3_bucket_public_access_block" "database_backups" {
  bucket = aws_s3_bucket.database_backups.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Lifecycle policy for database backups (retain for 90 days in prod, 30 in dev)
resource "aws_s3_bucket_lifecycle_configuration" "database_backups" {
  bucket = aws_s3_bucket.database_backups.id

  # Dev/Non-Prod Rule: Simple delete at 30 days (no transitions to avoid day=30 conflict)
  dynamic "rule" {
    for_each = var.environment != "prod" ? [1] : []
    content {
      id     = "30-day-delete"
      status = "Enabled"
      filter { prefix = "" }

      expiration {
        days = 30
      }
    }
  }

  # Prod Rule: Transitions + delete at 90 days
  dynamic "rule" {
    for_each = var.environment == "prod" ? [1] : []
    content {
      id     = "90-day-retention"
      status = "Enabled"
      filter { prefix = "" }

      transition {
        days          = 30
        storage_class = "STANDARD_IA"
      }

      transition {
        days          = 60
        storage_class = "GLACIER"
      }

      expiration {
        days = 90
      }
    }
  }
}
# S3 Bucket for Application Logs
resource "aws_s3_bucket" "logs" {
  bucket = "${var.project_name}-${var.environment}-logs"

  tags = {
    Name        = "${var.project_name}-${var.environment}-logs"
    Environment = var.environment
    Purpose     = "Application and access logs"
  }
}

# Server-side encryption for logs
resource "aws_s3_bucket_server_side_encryption_configuration" "logs" {
  bucket = aws_s3_bucket.logs.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# Block public access for logs
resource "aws_s3_bucket_public_access_block" "logs" {
  bucket = aws_s3_bucket.logs.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Lifecycle policy for logs (delete after retention period)
resource "aws_s3_bucket_lifecycle_configuration" "logs" {
  bucket = aws_s3_bucket.logs.id

  rule {
    id     = "delete-old-logs"
    status = "Enabled"
    filter {
      prefix = ""
    }

    expiration {
      days = var.log_retention_days
    }
  }
}

# IAM Policy for EC2 instances to access S3 buckets
resource "aws_iam_policy" "s3_access" {
  name        = "${var.project_name}-${var.environment}-s3-access"
  description = "Allow EC2 instances to access S3 buckets"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject",
          "s3:ListBucket"
        ]
        Resource = [
          aws_s3_bucket.jenkins_artifacts.arn,
          "${aws_s3_bucket.jenkins_artifacts.arn}/*",
          aws_s3_bucket.database_backups.arn,
          "${aws_s3_bucket.database_backups.arn}/*",
          aws_s3_bucket.logs.arn,
          "${aws_s3_bucket.logs.arn}/*"
        ]
      }
    ]
  })

  tags = {
    Name        = "${var.project_name}-${var.environment}-s3-policy"
    Environment = var.environment
  }
}