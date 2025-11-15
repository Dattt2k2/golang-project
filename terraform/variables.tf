# AWS Configuration
variable "aws_region" {
  description = "AWS region for resources"
  type        = string
  default     = "us-east-1"
}

variable "project_name" {
  description = "Project name for resource naming"
  type        = string
  default     = "microservices"
}

variable "environment" {
  description = "Environment (dev, staging, prod)"
  type        = string
  default     = "prod"
}

# VPC Configuration
variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "public_subnet_cidrs" {
  description = "CIDR blocks for public subnets"
  type        = list(string)
  default     = ["10.0.1.0/24", "10.0.2.0/24"]
}

variable "private_subnet_cidrs" {
  description = "CIDR blocks for private subnets"
  type        = list(string)
  default     = ["10.0.11.0/24", "10.0.12.0/24"]
}

# EC2 Configuration
variable "ec2_instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t3.large"
}

variable "ec2_ami" {
  description = "AMI ID for EC2 instance (Ubuntu 22.04 LTS)"
  type        = string
  default     = "" # Will use data source to find latest Ubuntu AMI
}

variable "ec2_key_pair_name" {
  description = "Name of SSH key pair for EC2 access"
  type        = string
  default     = "microservices-key"
}

variable "ec2_root_volume_size" {
  description = "Size of EC2 root volume in GB"
  type        = number
  default     = 50
}

# RDS Configuration
variable "rds_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t3.micro"
}

variable "rds_engine_version" {
  description = "PostgreSQL engine version"
  type        = string
  default     = "15.4"
}

variable "rds_allocated_storage" {
  description = "Allocated storage for RDS in GB"
  type        = number
  default     = 20
}

variable "rds_max_allocated_storage" {
  description = "Maximum allocated storage for RDS autoscaling in GB"
  type        = number
  default     = 100
}

variable "db_username" {
  description = "Master username for RDS databases"
  type        = string
  default     = "dbadmin"
  sensitive   = true
}

variable "db_password" {
  description = "Master password for RDS databases"
  type        = string
  sensitive   = true
}

# Database names
variable "auth_db_name" {
  description = "Auth service database name"
  type        = string
  default     = "authdb"
}

variable "user_db_name" {
  description = "User service database name"
  type        = string
  default     = "userdb"
}

variable "payment_db_name" {
  description = "Payment service database name"
  type        = string
  default     = "paymentdb"
}

variable "order_db_name" {
  description = "Order service database name"
  type        = string
  default     = "orderdb"
}

# Security
variable "allowed_ssh_cidr" {
  description = "CIDR blocks allowed to SSH into EC2"
  type        = list(string)
  default     = ["0.0.0.0/0"] # Restrict this to your IP in production!
}

variable "allowed_http_cidr" {
  description = "CIDR blocks allowed to access HTTP/HTTPS"
  type        = list(string)
  default     = ["0.0.0.0/0"]
}

# Application Configuration
variable "docker_compose_version" {
  description = "Docker Compose version to install"
  type        = string
  default     = "2.24.0"
}

variable "domain_name" {
  description = "Domain name for the application (optional)"
  type        = string
  default     = ""
}

# Tags
variable "tags" {
  description = "Common tags for all resources"
  type        = map(string)
  default = {
    Project     = "Microservices"
    ManagedBy   = "Terraform"
    Environment = "production"
  }
}
