terraform {
  required_version = ">= 1.0"
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  # Uncomment after creating S3 bucket for state management
  # backend "s3" {
  #   bucket         = "ecommerce-terraform-state"
  #   key            = "dev/terraform.tfstate"
  #   region         = "us-east-1"
  #   dynamodb_table = "ecommerce-terraform-locks"
  #   encrypt        = true
  # }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = var.project_name
      Environment = var.environment
      ManagedBy   = "Terraform"
    }
  }
}

# VPC Module
module "vpc" {
  source = "../../modules/vpc"

  project_name         = var.project_name
  environment          = var.environment
  vpc_cidr             = var.vpc_cidr
  public_subnet_cidrs  = var.public_subnet_cidrs
  private_subnet_cidrs = var.private_subnet_cidrs
  availability_zones   = var.availability_zones
}

# Security Groups Module
module "security" {
  source = "../../modules/security"

  project_name      = var.project_name
  environment       = var.environment
  vpc_id            = module.vpc.vpc_id
  allowed_ssh_cidrs = var.allowed_ssh_cidrs
}

# Storage Module (S3 Buckets)
module "storage" {
  source = "../../modules/storage"

  project_name            = var.project_name
  environment             = var.environment
  artifact_retention_days = var.artifact_retention_days
  backup_retention_days   = var.backup_retention_days
  log_retention_days      = var.log_retention_days
}

# Compute Module (EC2 + ALB)
module "compute" {
  source = "../../modules/compute"

  project_name          = var.project_name
  environment           = var.environment
  vpc_id                = module.vpc.vpc_id
  public_subnet_ids     = module.vpc.public_subnet_ids
  private_subnet_ids    = module.vpc.private_subnet_ids
  app_security_group_id = module.security.app_security_group_id
  alb_security_group_id = module.security.alb_security_group_id
  instance_type         = var.instance_type
  instance_count        = var.instance_count
  ssh_public_key        = file(var.ssh_public_key_path)
  s3_access_policy_arn  = module.storage.s3_access_policy_arn
  root_volume_size      = var.root_volume_size
}

# Database Module (RDS + Redis)
module "database" {
  source = "../../modules/database"

  project_name               = var.project_name
  environment                = var.environment
  private_subnet_ids         = module.vpc.private_subnet_ids
  database_security_group_id = module.security.database_security_group_id
  redis_security_group_id    = module.security.redis_security_group_id
  db_instance_class          = var.db_instance_class
  db_allocated_storage       = var.db_allocated_storage
  db_name                    = var.db_name
  db_username                = var.db_username
  db_password                = var.db_password
  redis_node_type            = var.redis_node_type
}