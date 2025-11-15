# AWS Microservices Deployment with Terraform

This directory contains Terraform configuration for deploying the microservices architecture on AWS using EC2 and RDS.

## Architecture Overview

- **EC2 Instance**: t3.large instance running Docker containers
- **RDS Databases**: 4 separate PostgreSQL databases (auth, user, payment, order)
- **VPC**: Custom VPC with public and private subnets
- **Elastic IP**: Static IP address for EC2 instance
- **Security Groups**: Configured for EC2 and RDS access

## Estimated Monthly Cost

- EC2 t3.large: ~$60/month
- RDS db.t3.micro x4: ~$15-20/month (free tier eligible)
- Storage and data transfer: ~$10-15/month
- **Total**: ~$85-95/month

## Prerequisites

1. **AWS Account** with appropriate permissions
2. **AWS CLI** configured with credentials
3. **Terraform** installed (v1.0+)
4. **SSH Key Pair** created in AWS EC2

## Setup Instructions

### 1. Configure AWS Credentials

```bash
aws configure
# Enter your AWS Access Key ID, Secret Access Key, and region
```

### 2. Create SSH Key Pair

```bash
# In AWS Console, create a key pair named "microservices-key"
# Or use AWS CLI:
aws ec2 create-key-pair --key-name microservices-key --query 'KeyMaterial' --output text > ~/.ssh/microservices-key.pem
chmod 400 ~/.ssh/microservices-key.pem
```

### 3. Update Configuration

Edit `terraform.tfvars`:
```hcl
# Change these values
db_password = "YourSecurePassword123!"
ec2_key_pair_name = "your-key-pair-name"
allowed_ssh_cidr = ["YOUR_IP/32"]  # Your IP address
```

### 4. Initialize Terraform

```bash
cd terraform
terraform init
```

### 5. Review Plan

```bash
terraform plan
```

### 6. Deploy Infrastructure

```bash
terraform apply
# Type "yes" when prompted
```

This will take 10-15 minutes to create all resources.

### 7. Get Outputs

```bash
terraform output
```

You'll see:
- EC2 public IP
- Database endpoints
- SSH command
- API Gateway URL

## Post-Deployment Setup

### 1. Connect to EC2

```bash
ssh -i ~/.ssh/microservices-key.pem ubuntu@<EC2_PUBLIC_IP>
```

### 2. Verify Installation

```bash
# Check Docker
docker --version
docker-compose --version

# Check services directory
ls -la /opt/microservices
```

### 3. Upload Application Files

From your local machine:

```bash
# Copy docker-compose.yaml
scp -i ~/.ssh/microservices-key.pem docker-compose.yaml ubuntu@<EC2_IP>:/opt/microservices/

# Copy other necessary files
scp -i ~/.ssh/microservices-key.pem -r traefik-config ubuntu@<EC2_IP>:/opt/microservices/
scp -i ~/.ssh/microservices-key.pem filebeat.yml ubuntu@<EC2_IP>:/opt/microservices/
```

Or use Git:

```bash
ssh -i ~/.ssh/microservices-key.pem ubuntu@<EC2_IP>
cd /opt/microservices
git clone <your-repo-url> .
```

### 4. Configure Environment Variables

```bash
ssh -i ~/.ssh/microservices-key.pem ubuntu@<EC2_IP>
cd /opt/microservices
nano .env
```

Add your secrets:
```env
# Stripe
STRIPE_SECRET_KEY=sk_live_...
STRIPE_WEBHOOK_SECRET=whsec_...

# JWT
JWT_SECRET=your-secure-jwt-secret

# Google OAuth
GOOGLE_CLIENT_ID=...
GOOGLE_CLIENT_SECRET=...

# AWS S3 (if using)
AWS_ACCESS_KEY_ID=...
AWS_SECRET_ACCESS_KEY=...
AWS_S3_BUCKET=...
```

### 5. Build and Deploy Services

```bash
# Deploy services
sudo ./deploy.sh

# Or manually:
sudo docker-compose pull
sudo docker-compose up -d

# Check status
sudo docker-compose ps

# View logs
sudo docker-compose logs -f
```

## Database Migration

The RDS databases are created but empty. Run migrations:

```bash
# SSH into EC2
ssh -i ~/.ssh/microservices-key.pem ubuntu@<EC2_IP>
cd /opt/microservices

# Run migrations for each service
sudo docker-compose exec auth-service ./migrate
sudo docker-compose exec user-service ./migrate
sudo docker-compose exec payment-service ./migrate
sudo docker-compose exec order-service ./migrate
```

## Accessing Services

- **API Gateway**: http://<EC2_IP>:8080
- **Traefik Dashboard**: http://<EC2_IP>:8081
- **Kibana**: http://<EC2_IP>:5601

## Monitoring

### CloudWatch Alarms

Configured alarms:
- EC2 CPU utilization > 80%
- EC2 status check failures
- RDS CPU utilization > 80%
- RDS free storage < 2GB

View in AWS Console → CloudWatch → Alarms

### Application Logs

```bash
# View all logs
sudo docker-compose logs -f

# View specific service
sudo docker-compose logs -f auth-service

# View Elasticsearch logs
curl http://localhost:9200/_cat/indices
```

## Backup and Restore

### RDS Automated Backups

- Retention: 7 days
- Backup window: 3:00-4:00 AM UTC
- Maintenance window: Monday 4:00-5:00 AM UTC

### Manual Backup

```bash
# Create RDS snapshot via AWS CLI
aws rds create-db-snapshot \
  --db-instance-identifier microservices-prod-auth-db \
  --db-snapshot-identifier manual-backup-$(date +%Y%m%d-%H%M%S)
```

### Database Restore

```bash
# Restore from snapshot
aws rds restore-db-instance-from-db-snapshot \
  --db-instance-identifier microservices-prod-auth-db-restored \
  --db-snapshot-identifier snapshot-name
```

## Scaling

### Vertical Scaling (Upgrade Instance)

```hcl
# In terraform.tfvars
ec2_instance_type = "t3.xlarge"  # 4 vCPUs, 16GB RAM
rds_instance_class = "db.t3.small"  # 2 vCPUs, 2GB RAM
```

```bash
terraform apply
```

### Horizontal Scaling (Add Instances)

For true horizontal scaling, consider upgrading to:
- Application Load Balancer + Auto Scaling Group
- Amazon ECS/Fargate
- Amazon EKS (Kubernetes)

## Security Best Practices

### 1. Restrict SSH Access

```hcl
# In terraform.tfvars
allowed_ssh_cidr = ["YOUR_IP/32"]
```

### 2. Enable MFA

Enable MFA for AWS root and IAM users.

### 3. Rotate Credentials

```bash
# Rotate database password
aws rds modify-db-instance \
  --db-instance-identifier microservices-prod-auth-db \
  --master-user-password "NewSecurePassword123!"

# Update .env file
nano /opt/microservices/.env
```

### 4. Enable VPN

For production, use AWS Client VPN or SSH tunnel:

```bash
# SSH tunnel to RDS
ssh -i ~/.ssh/microservices-key.pem -L 5432:auth-db-endpoint:5432 ubuntu@<EC2_IP>

# Connect to database via localhost:5432
psql -h localhost -U dbadmin -d authdb
```

### 5. Setup SSL/TLS

```bash
# Install Certbot
sudo snap install --classic certbot

# Get SSL certificate
sudo certbot --nginx -d api.yourdomain.com

# Configure Traefik for HTTPS
```

## Troubleshooting

### EC2 Issues

```bash
# Check instance status
aws ec2 describe-instance-status --instance-ids <instance-id>

# View system logs
aws ec2 get-console-output --instance-id <instance-id>

# SSH connection issues
chmod 400 ~/.ssh/microservices-key.pem
ssh -vvv -i ~/.ssh/microservices-key.pem ubuntu@<EC2_IP>
```

### RDS Issues

```bash
# Check database status
aws rds describe-db-instances --db-instance-identifier microservices-prod-auth-db

# Test database connection from EC2
sudo docker run --rm -it postgres:15 psql -h <db-endpoint> -U dbadmin -d authdb
```

### Docker Issues

```bash
# Restart Docker
sudo systemctl restart docker

# Clean up
sudo docker system prune -a

# Check disk space
df -h
```

## Cost Optimization

### 1. Use Reserved Instances

Save up to 72% with 1-3 year commitments:
```bash
# Purchase Reserved Instance via AWS Console
# EC2 → Reserved Instances → Purchase
```

### 2. Stop Non-Production Resources

```bash
# Stop EC2 (saves ~50% during stopped time)
aws ec2 stop-instances --instance-ids <instance-id>

# Stop RDS (saves ~100% during stopped time, max 7 days)
aws rds stop-db-instance --db-instance-identifier microservices-prod-auth-db
```

### 3. Use Spot Instances (Development)

For dev/test environments:
```hcl
resource "aws_spot_instance_request" "dev" {
  spot_price    = "0.05"
  instance_type = "t3.large"
  # ... other config
}
```

## Disaster Recovery

### Backup Strategy

1. **RDS Automated Backups**: 7 days retention
2. **Manual Snapshots**: Before major changes
3. **Application Data**: S3 backups
4. **Configuration**: Git repository

### Recovery Procedure

```bash
# 1. Restore from Terraform state
terraform apply

# 2. Restore RDS from snapshot
aws rds restore-db-instance-from-db-snapshot \
  --db-instance-identifier new-instance \
  --db-snapshot-identifier backup-snapshot

# 3. Update DNS/endpoints
# 4. Verify application connectivity
# 5. Test critical functions
```

## Updating Infrastructure

```bash
# Pull latest changes
git pull

# Review changes
terraform plan

# Apply updates
terraform apply

# Or target specific resources
terraform apply -target=aws_instance.main
```

## Destroying Infrastructure

⚠️ **WARNING**: This will delete all resources and data!

```bash
# Remove deletion protection from RDS
terraform apply -var="deletion_protection=false"

# Destroy all resources
terraform destroy

# Type "yes" to confirm
```

## Support

For issues or questions:
1. Check AWS CloudWatch logs
2. Review Terraform state: `terraform show`
3. Check AWS Service Health Dashboard
4. Contact AWS Support (if you have a support plan)

## Additional Resources

- [Terraform AWS Provider Docs](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- [AWS EC2 Documentation](https://docs.aws.amazon.com/ec2/)
- [AWS RDS Documentation](https://docs.aws.amazon.com/rds/)
- [Docker Documentation](https://docs.docker.com/)
