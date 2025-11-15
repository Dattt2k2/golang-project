# VPC Outputs
output "vpc_id" {
  description = "VPC ID"
  value       = aws_vpc.main.id
}

output "public_subnet_ids" {
  description = "Public subnet IDs"
  value       = aws_subnet.public[*].id
}

output "private_subnet_ids" {
  description = "Private subnet IDs"
  value       = aws_subnet.private[*].id
}

# EC2 Outputs
output "ec2_instance_id" {
  description = "EC2 instance ID"
  value       = aws_instance.main.id
}

output "ec2_public_ip" {
  description = "EC2 public IP address"
  value       = aws_eip.main.public_ip
}

output "ec2_private_ip" {
  description = "EC2 private IP address"
  value       = aws_instance.main.private_ip
}

# RDS Outputs
output "auth_db_endpoint" {
  description = "Auth database endpoint"
  value       = aws_db_instance.auth.endpoint
}

output "auth_db_address" {
  description = "Auth database address"
  value       = aws_db_instance.auth.address
}

output "user_db_endpoint" {
  description = "User database endpoint"
  value       = aws_db_instance.user.endpoint
}

output "user_db_address" {
  description = "User database address"
  value       = aws_db_instance.user.address
}

output "payment_db_endpoint" {
  description = "Payment database endpoint"
  value       = aws_db_instance.payment.endpoint
}

output "payment_db_address" {
  description = "Payment database address"
  value       = aws_db_instance.payment.address
}

output "order_db_endpoint" {
  description = "Order database endpoint"
  value       = aws_db_instance.order.endpoint
}

output "order_db_address" {
  description = "Order database address"
  value       = aws_db_instance.order.address
}

# Connection Information
output "ssh_command" {
  description = "SSH command to connect to EC2"
  value       = "ssh -i ~/.ssh/${var.ec2_key_pair_name}.pem ubuntu@${aws_eip.main.public_ip}"
}

output "api_gateway_url" {
  description = "API Gateway URL"
  value       = "http://${aws_eip.main.public_ip}:8080"
}

output "traefik_dashboard_url" {
  description = "Traefik Dashboard URL"
  value       = "http://${aws_eip.main.public_ip}:8081"
}

output "kibana_url" {
  description = "Kibana URL"
  value       = "http://${aws_eip.main.public_ip}:5601"
}

# Database Connection Strings
output "database_connections" {
  description = "Database connection information"
  value = {
    auth_db = {
      host     = aws_db_instance.auth.address
      port     = aws_db_instance.auth.port
      database = aws_db_instance.auth.db_name
      username = aws_db_instance.auth.username
    }
    user_db = {
      host     = aws_db_instance.user.address
      port     = aws_db_instance.user.port
      database = aws_db_instance.user.db_name
      username = aws_db_instance.user.username
    }
    payment_db = {
      host     = aws_db_instance.payment.address
      port     = aws_db_instance.payment.port
      database = aws_db_instance.payment.db_name
      username = aws_db_instance.payment.username
    }
    order_db = {
      host     = aws_db_instance.order.address
      port     = aws_db_instance.order.port
      database = aws_db_instance.order.db_name
      username = aws_db_instance.order.username
    }
  }
  sensitive = true
}

# Deployment Instructions
output "deployment_instructions" {
  description = "Next steps for deployment"
  value = <<-EOT
    
    ====================================
    Deployment Instructions
    ====================================
    
    1. Connect to EC2:
       ${self.ssh_command}
    
    2. Navigate to application directory:
       cd /opt/microservices
    
    3. Upload your docker-compose.yaml and application code
    
    4. Update .env file with your secrets:
       nano .env
    
    5. Deploy application:
       ./deploy.sh
    
    6. Check service status:
       docker-compose ps
    
    7. View logs:
       docker-compose logs -f
    
    API Gateway: ${self.api_gateway_url}
    Traefik Dashboard: ${self.traefik_dashboard_url}
    Kibana: ${self.kibana_url}
    
    ====================================
  EOT
}
