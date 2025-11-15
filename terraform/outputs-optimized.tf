# ALB Outputs
output "alb_dns_name" {
  description = "DNS name of the Application Load Balancer"
  value       = aws_lb.main.dns_name
}

output "alb_url" {
  description = "URL of the Application Load Balancer"
  value       = "http://${aws_lb.main.dns_name}"
}

# Traefik Instance
output "traefik_instance" {
  description = "Traefik API Gateway instance information"
  value = {
    id         = aws_instance.traefik.id
    public_ip  = aws_instance.traefik.public_ip
    private_ip = aws_instance.traefik.private_ip
  }
}

output "traefik_dashboard_url" {
  description = "Traefik Dashboard URL"
  value       = "http://${aws_instance.traefik.public_ip}:8080"
}

# Service Instances
output "service_instances" {
  description = "Microservice instance information"
  value = {
    for k, v in aws_instance.services : k => {
      id         = v.id
      public_ip  = v.public_ip
      private_ip = v.private_ip
    }
  }
}

# Shared Infrastructure Instance
output "shared_infrastructure" {
  description = "Shared infrastructure instance information"
  value = {
    id         = aws_instance.shared_infrastructure.id
    private_ip = aws_instance.shared_infrastructure.private_ip
  }
}

# Database Endpoints
output "database_endpoints" {
  description = "RDS database endpoints"
  value = {
    auth_db = {
      endpoint = aws_db_instance.auth.endpoint
      address  = aws_db_instance.auth.address
    }
    user_db = {
      endpoint = aws_db_instance.user.endpoint
      address  = aws_db_instance.user.address
    }
    payment_db = {
      endpoint = aws_db_instance.payment.endpoint
      address  = aws_db_instance.payment.address
    }
    order_db = {
      endpoint = aws_db_instance.order.endpoint
      address  = aws_db_instance.order.address
    }
  }
  sensitive = true
}

# SSH Commands
output "ssh_commands" {
  description = "SSH commands to connect to instances"
  value = {
    traefik = "ssh -i ~/.ssh/${var.ec2_key_pair_name}.pem ubuntu@${aws_instance.traefik.public_ip}"
    services = {
      for k, v in aws_instance.services : k => "ssh -i ~/.ssh/${var.ec2_key_pair_name}.pem ubuntu@${v.public_ip}"
    }
    infrastructure = "ssh -i ~/.ssh/${var.ec2_key_pair_name}.pem ubuntu@${aws_instance.shared_infrastructure.private_ip} # Via bastion"
  }
}

# Deployment Summary
output "deployment_summary" {
  description = "Optimized deployment summary"
  value = <<-EOT
    
    ====================================
    Optimized Multi-Instance Deployment
    ====================================
    
    Application Load Balancer:
    URL: http://${aws_lb.main.dns_name}
    
    API Gateway (Traefik):
    Public IP: ${aws_instance.traefik.public_ip}
    Dashboard: http://${aws_instance.traefik.public_ip}:8080
    
    Microservices:
    ${join("\n    ", [for k, v in aws_instance.services : "${k}: ${v.public_ip} (${v.private_ip})"])}
    
    Shared Infrastructure (Private):
    IP: ${aws_instance.shared_infrastructure.private_ip}
    Services: Redis:6379, Kafka:9092, Elasticsearch:9200, Kibana:5601
    
    RDS Databases:
    Auth DB: ${aws_db_instance.auth.endpoint}
    User DB: ${aws_db_instance.user.endpoint}
    Payment DB: ${aws_db_instance.payment.endpoint}
    Order DB: ${aws_db_instance.order.endpoint}
    
    Architecture Summary:
    - 1x ALB (Application Load Balancer)
    - 1x t3.small (Traefik Gateway)
    - 6x t3.micro (Microservices)
    - 1x t3.large (Shared Infrastructure)
    - 4x db.t3.micro (RDS PostgreSQL)
    
    Total: 8 EC2 + 4 RDS + 1 ALB
    
    Estimated Monthly Cost:
    - ALB: $16
    - Traefik (t3.small): $15
    - Services (6x t3.micro): $36
    - Infrastructure (t3.large): $60
    - RDS (4x db.t3.micro): $20
    - Storage & Data Transfer: $15
    - Total: ~$162/month
    
    Cost Savings vs Full Multi-Instance:
    $186 → $162 (Save $24/month or 13%)
    
    ====================================
  EOT
}

# Service URLs via Traefik
output "service_urls" {
  description = "Service URLs via Traefik"
  value = {
    auth    = "http://${aws_lb.main.dns_name}/api/auth"
    user    = "http://${aws_lb.main.dns_name}/api/user"
    product = "http://${aws_lb.main.dns_name}/api/product"
    cart    = "http://${aws_lb.main.dns_name}/api/cart"
    order   = "http://${aws_lb.main.dns_name}/api/order"
    payment = "http://${aws_lb.main.dns_name}/api/payment"
  }
}
