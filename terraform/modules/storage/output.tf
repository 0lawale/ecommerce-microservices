output "jenkins_artifacts_bucket_name" {
  description = "Name of the Jenkins artifacts S3 bucket"
  value       = aws_s3_bucket.jenkins_artifacts.id
}

output "jenkins_artifacts_bucket_arn" {
  description = "ARN of the Jenkins artifacts S3 bucket"
  value       = aws_s3_bucket.jenkins_artifacts.arn
}

output "database_backups_bucket_name" {
  description = "Name of the database backups S3 bucket"
  value       = aws_s3_bucket.database_backups.id
}

output "database_backups_bucket_arn" {
  description = "ARN of the database backups S3 bucket"
  value       = aws_s3_bucket.database_backups.arn
}

output "logs_bucket_name" {
  description = "Name of the logs S3 bucket"
  value       = aws_s3_bucket.logs.id
}

output "logs_bucket_arn" {
  description = "ARN of the logs S3 bucket"
  value       = aws_s3_bucket.logs.arn
}

output "s3_access_policy_arn" {
  description = "ARN of the IAM policy for S3 access"
  value       = aws_iam_policy.s3_access.arn
}