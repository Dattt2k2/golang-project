# Multi-Instance AWS Deployment

Kiến trúc này deploy mỗi microservice trên một EC2 instance riêng biệt với Application Load Balancer.

## Kiến Trúc

### Application Tier (Public Subnet)
- **API Gateway**: 1x t3.small instance
- **Auth Service**: 1x t3.micro instance
- **User Service**: 1x t3.micro instance
- **Product Service**: 1x t3.micro instance
- **Order Service**: 1x t3.micro instance
- **Payment Service**: 1x t3.micro instance

### Infrastructure Tier (Private Subnet)
- **Redis**: 1x t3.medium instance
- **Kafka**: 1x t3.medium instance
- **Elasticsearch + Kibana**: 1x t3.medium instance

### Database Tier (Private Subnet)
- **Auth DB**: RDS PostgreSQL db.t3.micro
- **User DB**: RDS PostgreSQL db.t3.micro
- **Payment DB**: RDS PostgreSQL db.t3.micro
- **Order DB**: RDS PostgreSQL db.t3.micro

### Load Balancing
- **Application Load Balancer** với path-based routing

## Chi Phí Ước Tính

| Component | Type | Quantity | Cost/Month |
|-----------|------|----------|------------|
| ALB | Application Load Balancer | 1 | $16 |
| API Gateway | t3.small | 1 | $15 |
| Microservices | t3.micro | 5 | $30 |
| Infrastructure | t3.medium | 3 | $90 |
| RDS | db.t3.micro | 4 | $20 |
| Storage & Transfer | - | - | $15 |
| **TỔNG** | | | **~$186/month** |

## Ưu Điểm

✅ **Isolation**: Mỗi service độc lập, lỗi ở service này không ảnh hưởng service khác
✅ **Scalability**: Dễ dàng scale từng service riêng lẻ
✅ **Monitoring**: Theo dõi tài nguyên từng service rõ ràng
✅ **Deployment**: Deploy/rollback từng service độc lập
✅ **High Availability**: ALB tự động health check và routing
✅ **Security**: Infrastructure services ở private subnet

## Nhược Điểm

❌ **Cost**: Đắt hơn single-instance (~$186 vs $95/month)
❌ **Complexity**: Phức tạp hơn trong quản lý
❌ **Network Latency**: Inter-service communication qua network

## So Sánh Với Single-Instance

| Aspect | Single Instance | Multi-Instance |
|--------|----------------|----------------|
| Cost | $85-95/month | $186/month |
| Instances | 1 EC2 | 9 EC2 + ALB |
| Isolation | ❌ Shared resources | ✅ Isolated |
| Scalability | ❌ Vertical only | ✅ Per-service |
| High Availability | ❌ Single point of failure | ✅ ALB + multiple instances |
| Complexity | ✅ Simple | ❌ Complex |

## Deployment

### 1. Sử dụng multi-instance configuration

```bash
cd terraform

# Rename files to use multi-instance configs
mv ec2.tf ec2-single.tf.bak
mv security_groups.tf security_groups-single.tf.bak
mv outputs.tf outputs-single.tf.bak
mv variables.tf variables-single.tf.bak
mv terraform.tfvars terraform-single.tfvars.bak

# Use multi-instance configs
mv ec2-multi-instance.tf ec2.tf
mv security_groups-multi-instance.tf security_groups.tf
mv outputs-multi-instance.tf outputs.tf
mv variables-multi-instance.tf variables-append.tf
mv terraform-multi-instance.tfvars terraform.tfvars

# Append additional variables
cat variables-append.tf >> variables.tf
```

### 2. Cấu hình

```bash
# Edit terraform.tfvars
nano terraform.tfvars
```

Thay đổi:
- `db_password`: Mật khẩu mạnh
- `ec2_key_pair_name`: Tên SSH key của bạn
- `allowed_ssh_cidr`: IP của bạn

### 3. Deploy

```bash
# Initialize
terraform init

# Review plan
terraform plan

# Deploy (15-20 phút)
terraform apply
```

### 4. Lấy thông tin

```bash
# ALB URL
terraform output alb_dns_name

# Service IPs
terraform output service_instances

# Infrastructure IPs
terraform output infrastructure_instances

# SSH commands
terraform output service_ssh_commands
```

## Truy Cập Services

### Qua ALB (Production)
```bash
# API Gateway (default)
http://<ALB-DNS>/

# Specific services
http://<ALB-DNS>/api/auth/*
http://<ALB-DNS>/api/user/*
http://<ALB-DNS>/api/product/*
http://<ALB-DNS>/api/order/*
http://<ALB-DNS>/api/payment/*
```

### Direct Access (Debugging)
```bash
# SSH vào từng service
ssh -i ~/.ssh/microservices-key.pem ubuntu@<SERVICE-PUBLIC-IP>

# Check service status
sudo systemctl status <service-name>

# View logs
sudo journalctl -u <service-name> -f
```

## Auto Scaling (Tùy Chọn)

Để thêm auto-scaling, tạo file `autoscaling.tf`:

```hcl
resource "aws_autoscaling_group" "auth_service" {
  name                = "${var.project_name}-${var.environment}-auth-asg"
  vpc_zone_identifier = aws_subnet.public[*].id
  target_group_arns   = [aws_lb_target_group.auth_service.arn]
  health_check_type   = "ELB"
  
  min_size         = 1
  max_size         = 3
  desired_capacity = 1
  
  launch_template {
    id      = aws_launch_template.auth_service.id
    version = "$Latest"
  }
}

resource "aws_autoscaling_policy" "auth_cpu" {
  name                   = "${var.project_name}-${var.environment}-auth-cpu-scaling"
  autoscaling_group_name = aws_autoscaling_group.auth_service.name
  policy_type            = "TargetTrackingScaling"
  
  target_tracking_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ASGAverageCPUUtilization"
    }
    target_value = 70.0
  }
}
```

## Monitoring

### CloudWatch Alarms
Tự động tạo alarms cho:
- CPU utilization > 80%
- Status check failures
- Target group health

### View Metrics
```bash
# CPU usage của tất cả services
aws cloudwatch get-metric-statistics \
  --namespace AWS/EC2 \
  --metric-name CPUUtilization \
  --dimensions Name=InstanceId,Value=<instance-id> \
  --start-time $(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%S) \
  --end-time $(date -u +%Y-%m-%dT%H:%M:%S) \
  --period 300 \
  --statistics Average
```

## High Availability Setup

### Multi-AZ Deployment
Chỉnh sửa `ec2.tf`:

```hcl
resource "aws_instance" "services" {
  for_each = local.services

  # Distribute across multiple AZs
  subnet_id = element(aws_subnet.public[*].id, index(keys(local.services), each.key))
  
  # ... rest of config
}
```

### Health Checks
ALB tự động health check mỗi 30s:
- Path: `/health`
- Timeout: 5s
- Healthy threshold: 2 checks
- Unhealthy threshold: 2 checks

## SSL/TLS Setup

### 1. Request Certificate (ACM)
```bash
aws acm request-certificate \
  --domain-name api.yourdomain.com \
  --validation-method DNS
```

### 2. Update ALB Listener
```hcl
resource "aws_lb_listener" "https" {
  load_balancer_arn = aws_lb.main.arn
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = "arn:aws:acm:region:account:certificate/xxx"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.api_gateway.arn
  }
}

# Redirect HTTP to HTTPS
resource "aws_lb_listener" "http_redirect" {
  load_balancer_arn = aws_lb.main.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type = "redirect"
    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}
```

## Backup Strategy

### Application State
```bash
# SSH vào từng service và backup
for service in auth user product order payment; do
  ssh ubuntu@<IP> "sudo systemctl stop $service-service"
  ssh ubuntu@<IP> "sudo tar -czf /tmp/$service-backup.tar.gz /opt/$service-service"
  scp ubuntu@<IP>:/tmp/$service-backup.tar.gz ./backups/
  ssh ubuntu@<IP> "sudo systemctl start $service-service"
done
```

### Infrastructure Data
```bash
# Redis backup
ssh ubuntu@<REDIS-IP> "sudo docker exec redis redis-cli BGSAVE"
ssh ubuntu@<REDIS-IP> "sudo docker cp redis:/data/dump.rdb /tmp/"
scp ubuntu@<REDIS-IP>:/tmp/dump.rdb ./backups/redis/

# Kafka topics backup (if needed)
ssh ubuntu@<KAFKA-IP> "sudo docker exec kafka kafka-consumer-groups --bootstrap-server localhost:9092 --list"
```

## Troubleshooting

### Service không healthy
```bash
# Check target group health
aws elbv2 describe-target-health \
  --target-group-arn <target-group-arn>

# SSH vào instance và check
ssh ubuntu@<IP>
sudo systemctl status <service>
sudo journalctl -u <service> -n 100
curl localhost:<port>/health
```

### ALB không route
```bash
# Check ALB rules
aws elbv2 describe-rules --listener-arn <listener-arn>

# Check security groups
aws ec2 describe-security-groups --group-ids <sg-id>
```

### Inter-service communication failed
```bash
# Test từ service instance
ssh ubuntu@<SERVICE-IP>
curl http://<INFRASTRUCTURE-PRIVATE-IP>:6379  # Redis
curl http://<INFRASTRUCTURE-PRIVATE-IP>:9200  # Elasticsearch
```

## Cost Optimization

### 1. Use Spot Instances (Dev/Test)
```hcl
resource "aws_instance" "services" {
  instance_market_options {
    market_type = "spot"
    spot_options {
      max_price = "0.01"  # $0.01/hour
    }
  }
}
```
Tiết kiệm: ~70%

### 2. Reserved Instances (Production)
Commit 1 năm để tiết kiệm 40-72%

### 3. Schedule Stop/Start
```bash
# Stop non-production instances at night
aws ec2 stop-instances --instance-ids i-xxx i-yyy
```

### 4. Right-sizing
Monitor và adjust instance types dựa trên actual usage

## Migration từ Single-Instance

### 1. Export data từ single instance
```bash
ssh ubuntu@<SINGLE-INSTANCE-IP>
sudo docker-compose -f docker-compose.prod.yaml exec redis redis-cli BGSAVE
# Export databases, configs, etc.
```

### 2. Deploy multi-instance infrastructure
```bash
terraform apply
```

### 3. Import data vào new instances
```bash
# Import vào từng service
scp backups/redis/dump.rdb ubuntu@<REDIS-IP>:/tmp/
ssh ubuntu@<REDIS-IP> "sudo docker cp /tmp/dump.rdb redis:/data/"
```

### 4. Update DNS
Point domain to ALB DNS name

### 5. Verify và destroy old instance
```bash
# Test thoroughly
curl http://<ALB-DNS>/health

# Destroy old instance
terraform destroy -target=aws_instance.main
```

## Khi Nào Nên Dùng?

**Nên dùng Multi-Instance khi:**
- Production environment cần high availability
- Cần scale từng service riêng lẻ
- Budget cho phép (~$186/month)
- Team đủ kinh nghiệm quản lý infrastructure phức tạp

**Nên dùng Single-Instance khi:**
- Development/staging environment
- Startup với budget hạn chế
- Traffic thấp (< 1000 requests/minute)
- Ưu tiên đơn giản hơn chi phí
