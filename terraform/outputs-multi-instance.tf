# ALB Outputs
output "alb_dns_name" {
  description = "DNS name of the Application Load Balancer"
  value       = aws_lb.main.dns_name
}

output "alb_url" {
  description = "URL of the Application Load Balancer"
  value       = "http://${aws_lb.main.dns_name}"
}

output "alb_zone_id" {
  description = "Zone ID of the ALB (for Route53)"
  value       = aws_lb.main.zone_id
}

# Service Instance IPs
output "service_instances" {
  description = "Service instance information"
  value = {
    for k, v in aws_instance.services : k => {
      id         = v.id
      public_ip  = v.public_ip
      private_ip = v.private_ip
    }
  }
}

# Infrastructure Instance IPs
output "infrastructure_instances" {
  description = "Infrastructure instance information"
  value = {
    for k, v in aws_instance.infrastructure : k => {
      id         = v.id
      private_ip = v.private_ip
    }
  }
}

# Connection Commands
output "service_ssh_commands" {
  description = "SSH commands to connect to service instances"
  value = {
    for k, v in aws_instance.services : k => "ssh -i ~/.ssh/${var.ec2_key_pair_name}.pem ubuntu@${v.public_ip}"
  }
}

# Target Group ARNs
output "target_group_arns" {
  description = "Target Group ARNs"
  value = {
    api_gateway     = aws_lb_target_group.api_gateway.arn
    auth_service    = aws_lb_target_group.auth_service.arn
    user_service    = aws_lb_target_group.user_service.arn
    product_service = aws_lb_target_group.product_service.arn
    order_service   = aws_lb_target_group.order_service.arn
    payment_service = aws_lb_target_group.payment_service.arn
  }
}

# Deployment Summary
output "deployment_summary" {
  description = "Deployment summary"
  value = <<-EOT
    
    ====================================
    Multi-Instance Deployment Summary
    ====================================
    
    Application Load Balancer:
    URL: http://${aws_lb.main.dns_name}
    
    Service Instances:
    ${join("\n    ", [for k, v in aws_instance.services : "${k}: ${v.public_ip}"])}
    
    Infrastructure Instances (Private):
    ${join("\n    ", [for k, v in aws_instance.infrastructure : "${k}: ${v.private_ip}"])}
    
    Database Endpoints:
    Auth DB: ${aws_db_instance.auth.endpoint}
    User DB: ${aws_db_instance.user.endpoint}
    Payment DB: ${aws_db_instance.payment.endpoint}
    Order DB: ${aws_db_instance.order.endpoint}
    
    Total EC2 Instances: ${length(aws_instance.services) + length(aws_instance.infrastructure)}
    Total RDS Instances: 4
    
    Estimated Monthly Cost:
    - ALB: ~$16
    - EC2 Services (6x t3.micro): ~$36
    - EC2 Infrastructure (3x t3.medium): ~$90
    - RDS (4x db.t3.micro): ~$20
    - Data Transfer & Storage: ~$15
    - Total: ~$177/month
    
    ====================================
  EOT
}
