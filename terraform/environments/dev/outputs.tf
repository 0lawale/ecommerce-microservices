# VPC Outputs
output "vpc_id" {
  description = "ID of the VPC"
  value       = module.vpc.vpc_id
}

output "public_subnet_ids" {
  description = "IDs of public subnets"
  value       = module.vpc.public_subnet_ids
}

output "private_subnet_ids" {
  description = "IDs of private subnets"
  value       = module.vpc.private_subnet_ids
}

# Compute Outputs
output "k3s_master_private_ip" {
  description = "Private IP of k3s master node"
  value       = module.compute.k3s_master_private_ip
}

output "k3s_worker_private_ips" {
  description = "Private IPs of k3s worker nodes"
  value       = module.compute.k3s_worker_private_ips
}

output "k3s_master_id" {
  description = "Instance ID of k3s master"
  value       = module.compute.k3s_master_id
}

output "k3s_worker_ids" {
  description = "Instance IDs of k3s workers"
  value       = module.compute.k3s_worker_ids
}

output "alb_dns_name" {
  description = "DNS name of the Application Load Balancer"
  value       = module.compute.alb_dns_name
}

# Database Outputs
output "rds_endpoint" {
  description = "RDS PostgreSQL endpoint"
  value       = module.database.postgres_endpoint
  sensitive   = true
}

output "rds_database_name" {
  description = "RDS database name"
  value       = module.database.postgres_database_name
}

output "redis_endpoint" {
  description = "ElastiCache Redis endpoint"
  value       = module.database.redis_endpoint
  sensitive   = true
}

# Storage Outputs
output "jenkins_artifacts_bucket" {
  description = "S3 bucket for Jenkins artifacts"
  value       = module.storage.jenkins_artifacts_bucket_name
}

output "database_backups_bucket" {
  description = "S3 bucket for database backups"
  value       = module.storage.database_backups_bucket_name
}

output "logs_bucket" {
  description = "S3 bucket for logs"
  value       = module.storage.logs_bucket_name
}

# Connection Info
output "connect_to_master" {
  description = "Command to connect to k3s master via SSM"
  value       = "aws ssm start-session --target ${module.compute.k3s_master_id}"
}

output "kubeconfig_command" {
  description = "Command to get kubeconfig from master (run after k3s is installed)"
  value       = "aws ssm start-session --target ${module.compute.k3s_master_id} --document-name AWS-StartInteractiveCommand --parameters command='sudo cat /etc/rancher/k3s/k3s.yaml'"
}