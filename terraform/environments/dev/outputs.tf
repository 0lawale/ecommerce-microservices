# VPC Outputs
output "vpc_id" {
  description = "VPC ID"
  value       = module.vpc.vpc_id
}

output "public_subnet_ids" {
  description = "Public subnet IDs"
  value       = module.vpc.public_subnet_ids
}

output "private_subnet_ids" {
  description = "Private subnet IDs"
  value       = module.vpc.private_subnet_ids
}

# Compute Outputs
output "alb_dns_name" {
  description = "Application Load Balancer DNS name"
  value       = module.compute.alb_dns_name
}

output "instance_private_ips" {
  description = "EC2 instance private IPs"
  value       = module.compute.instance_private_ips
}

# Database Outputs
output "postgres_endpoint" {
  description = "PostgreSQL endpoint"
  value       = module.database.postgres_endpoint
}

output "postgres_address" {
  description = "PostgreSQL address"
  value       = module.database.postgres_address
}

output "redis_endpoint" {
  description = "Redis endpoint"
  value       = module.database.redis_endpoint
}

# Connection strings (for your application)
output "database_connection_string" {
  description = "PostgreSQL connection string"
  value       = "postgresql://${var.db_username}:****@${module.database.postgres_address}:5432/${var.db_name}"
  sensitive   = true
}
