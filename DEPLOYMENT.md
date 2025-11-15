# Complete Terraform configuration files have been created in the terraform/ directory:

terraform/
├── main.tf              # Main Terraform configuration with provider setup
├── variables.tf         # All configurable variables
├── terraform.tfvars     # Default values for variables (customize this!)
├── vpc.tf              # VPC, subnets, route tables, internet gateway
├── security_groups.tf  # Security groups for EC2 and RDS
├── ec2.tf              # EC2 instance, Elastic IP, CloudWatch alarms
├── rds.tf              # 4 RDS PostgreSQL databases with backups
├── outputs.tf          # Output values after deployment
├── user_data.sh        # EC2 bootstrap script
└── README.md           # Comprehensive deployment guide

# Additional files created:
- docker-compose.prod.yaml  # Production Docker Compose configuration
- deploy.sh                 # Automated deployment script
- backup.sh                 # Database backup script
- restore.sh                # Database restore script

## Quick Start Guide

### 1. Prerequisites
- AWS Account with CLI configured
- Terraform installed (v1.0+)
- SSH key pair created in AWS

### 2. Configure Terraform
Edit terraform/terraform.tfvars:
```bash
cd terraform
nano terraform.tfvars
```

Update these critical values:
- db_password: Use a strong password
- ec2_key_pair_name: Your SSH key name
- allowed_ssh_cidr: Your IP address for security

### 3. Deploy Infrastructure
```bash
# Initialize Terraform
terraform init

# Review the plan
terraform plan

# Deploy (takes 10-15 minutes)
terraform apply
```

### 4. Get Connection Info
```bash
# View all outputs
terraform output

# Get EC2 IP
terraform output ec2_public_ip

# Get database endpoints
terraform output database_connections
```

### 5. Connect to EC2
```bash
# SSH into the server
ssh -i ~/.ssh/your-key.pem ubuntu@<EC2_IP>

# Navigate to app directory
cd /opt/microservices
```

### 6. Deploy Application
```bash
# Upload docker-compose.prod.yaml
scp -i ~/.ssh/your-key.pem docker-compose.prod.yaml ubuntu@<EC2_IP>:/opt/microservices/

# Upload deployment script
scp -i ~/.ssh/your-key.pem deploy.sh ubuntu@<EC2_IP>:/opt/microservices/

# SSH and deploy
ssh -i ~/.ssh/your-key.pem ubuntu@<EC2_IP>
cd /opt/microservices

# Update .env with your secrets
nano .env

# Run deployment
sudo ./deploy.sh
```

### 7. Verify Deployment
```bash
# Check all services
sudo docker-compose -f docker-compose.prod.yaml ps

# View logs
sudo docker-compose -f docker-compose.prod.yaml logs -f

# Test API
curl http://localhost:8080/health
```

## Architecture Details

### EC2 Instance
- Type: t3.large (2 vCPUs, 8GB RAM)
- Storage: 50GB GP3 SSD
- OS: Ubuntu 22.04 LTS
- Cost: ~$60/month

### RDS Databases (PostgreSQL 15.4)
1. auth-db: Authentication service
2. user-db: User management
3. payment-db: Payment processing
4. order-db: Order management

- Instance: db.t3.micro (free tier eligible)
- Storage: 20GB with autoscaling to 100GB
- Backups: 7-day retention
- Cost: ~$15-20/month total

### Networking
- VPC: 10.0.0.0/16
- Public Subnets: 10.0.1.0/24, 10.0.2.0/24 (EC2)
- Private Subnets: 10.0.11.0/24, 10.0.12.0/24 (RDS)
- Static IP: Elastic IP attached to EC2

### Security
- EC2 Security Group: SSH (22), HTTP (80), HTTPS (443), API (8080)
- RDS Security Group: PostgreSQL (5432) from EC2 only
- Encrypted storage for both EC2 and RDS
- VPN recommended for production database access

## Cost Optimization

### Current Setup: ~$85-95/month
- EC2 t3.large: $60
- RDS 4x db.t3.micro: $20
- Storage & networking: $10-15

### Savings Options:
1. **Reserved Instances**: Save up to 72% with 1-year commitment
2. **Smaller RDS**: Use db.t3.micro free tier (first 12 months)
3. **Stop when not needed**: Stop EC2/RDS during non-business hours
4. **Spot Instances**: For dev/test (up to 90% savings)

## Monitoring & Alerts

### CloudWatch Alarms Created:
- EC2 CPU > 80%
- EC2 status check failures
- RDS CPU > 80%
- RDS storage < 2GB

### View Logs:
- Application: `docker-compose logs -f`
- Elasticsearch: http://<EC2_IP>:9200
- Kibana: http://<EC2_IP>:5601

## Backup & Recovery

### Automated Backups:
- RDS: 7-day retention, daily at 3AM UTC
- Docker volumes: Use backup.sh script

### Manual Backup:
```bash
# Run backup script
./backup.sh

# Backups stored in ./backups/
# Automatically compressed and cleaned (7-day retention)
```

### Restore:
```bash
# List available backups
ls -lh ./backups/

# Restore specific database
./restore.sh auth ./backups/auth_20240101_120000.sql.gz
```

## Scaling

### Vertical Scaling:
Edit terraform/terraform.tfvars:
```hcl
ec2_instance_type = "t3.xlarge"  # 4 vCPUs, 16GB RAM
rds_instance_class = "db.t3.small"  # 2 vCPUs, 2GB RAM
```

Apply changes:
```bash
terraform apply
```

### Horizontal Scaling:
For production at scale, consider:
- Application Load Balancer + Auto Scaling Group
- Amazon ECS/Fargate for container orchestration
- Amazon EKS for Kubernetes
- Amazon Aurora for database clustering

## Security Checklist

- [ ] Change default database password
- [ ] Restrict SSH access to your IP only
- [ ] Setup VPN for database access
- [ ] Enable MFA on AWS account
- [ ] Rotate credentials regularly
- [ ] Setup SSL/TLS certificates
- [ ] Configure firewall rules
- [ ] Enable AWS CloudTrail
- [ ] Setup backup verification
- [ ] Document disaster recovery plan

## Troubleshooting

### Can't connect to EC2:
```bash
# Check security group allows your IP
# Verify key permissions
chmod 400 ~/.ssh/your-key.pem
```

### Database connection failed:
```bash
# Test from EC2
docker run --rm -it postgres:15 psql -h <db-endpoint> -U dbadmin -d authdb
```

### Services not starting:
```bash
# Check logs
docker-compose -f docker-compose.prod.yaml logs <service>

# Check disk space
df -h

# Check memory
free -h
```

## Next Steps

1. **Domain Setup**: Configure Route 53 or your DNS provider
2. **SSL/TLS**: Install Let's Encrypt certificates
3. **CI/CD**: Setup GitHub Actions or GitLab CI
4. **Monitoring**: Configure CloudWatch dashboards
5. **Alerting**: Setup SNS for alarm notifications

## Support

For issues:
1. Check logs: `docker-compose logs`
2. Review Terraform output: `terraform show`
3. Check AWS Service Health Dashboard
4. Review README.md in terraform/ directory

## Estimated Deployment Time

- Terraform apply: 10-15 minutes
- EC2 bootstrap: 5-10 minutes
- Application deployment: 5-10 minutes
- Total: ~20-35 minutes for full setup
