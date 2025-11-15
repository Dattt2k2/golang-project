# Optimized Multi-Instance Deployment Configuration

# AWS Region
aws_region = "us-east-1"

# Project Configuration
project_name = "microservices"
environment  = "prod"

# SSH Key Pair
ec2_key_pair_name = "microservices-key"

# EC2 Instance Types
gateway_instance_type               = "t3.small"  # Traefik (2 vCPU, 2GB RAM) - $15/month
service_instance_type               = "t3.micro"  # Microservices (2 vCPU, 1GB RAM) - $6/month each
shared_infrastructure_instance_type = "t3.large"  # Redis+Kafka+ES (2 vCPU, 8GB RAM) - $60/month

# RDS Configuration
rds_instance_class        = "db.t3.micro"
rds_allocated_storage     = 20
rds_max_allocated_storage = 100

# Database Credentials
db_username = "dbadmin"
db_password = "ChangeThisPassword123!" # Use a strong password

# Database Names
auth_db_name    = "authdb"
user_db_name    = "userdb"
payment_db_name = "paymentdb"
order_db_name   = "orderdb"

# Security Configuration
allowed_ssh_cidr  = ["0.0.0.0/0"] # Change to your IP: ["YOUR_IP/32"]
allowed_http_cidr = ["0.0.0.0/0"]

# Tags
tags = {
  Project      = "Microservices"
  ManagedBy    = "Terraform"
  Environment  = "production"
  Architecture = "Optimized-Multi-Instance"
  Team         = "DevOps"
}
