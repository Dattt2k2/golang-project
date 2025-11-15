# AWS Region
aws_region = "us-east-1"

# Project Configuration
project_name = "microservices"
environment  = "prod"

# EC2 Configuration
ec2_instance_type    = "t3.large"
ec2_key_pair_name    = "microservices-key" # Change this to your key pair name
ec2_root_volume_size = 50

# RDS Configuration
rds_instance_class        = "db.t3.micro"
rds_allocated_storage     = 20
rds_max_allocated_storage = 100

# Database Credentials (CHANGE THESE!)
db_username = "dbadmin"
db_password = "ChangeThisPassword123!" # Use a strong password

# Database Names
auth_db_name    = "authdb"
user_db_name    = "userdb"
payment_db_name = "paymentdb"
order_db_name   = "orderdb"

# Security Configuration
# IMPORTANT: Restrict these CIDR blocks to your IP address in production!
allowed_ssh_cidr  = ["0.0.0.0/0"] # Change to your IP: ["YOUR_IP/32"]
allowed_http_cidr = ["0.0.0.0/0"]

# Docker Configuration
docker_compose_version = "2.24.0"

# Optional: Domain Name
# domain_name = "api.yourdomain.com"

# Tags
tags = {
  Project     = "Microservices"
  ManagedBy   = "Terraform"
  Environment = "production"
  Team        = "DevOps"
}
