# Additional variables for optimized multi-instance deployment

# EC2 Instance Types
variable "gateway_instance_type" {
  description = "EC2 instance type for Traefik API Gateway"
  type        = string
  default     = "t3.small"
}

variable "service_instance_type" {
  description = "EC2 instance type for microservices"
  type        = string
  default     = "t3.micro"
}

variable "shared_infrastructure_instance_type" {
  description = "EC2 instance type for shared infrastructure (Redis, Kafka, Elasticsearch)"
  type        = string
  default     = "t3.large"
}

# ALB Configuration
variable "enable_alb_access_logs" {
  description = "Enable ALB access logs"
  type        = bool
  default     = false
}

variable "alb_access_logs_bucket" {
  description = "S3 bucket for ALB access logs"
  type        = string
  default     = ""
}
