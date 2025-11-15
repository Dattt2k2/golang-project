# Get latest Ubuntu 22.04 LTS AMI
data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"] # Canonical

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

# Elastic IP for EC2
resource "aws_eip" "main" {
  domain = "vpc"
  
  tags = merge(var.tags, {
    Name = "${var.project_name}-${var.environment}-eip"
  })
}

# EC2 Instance
resource "aws_instance" "main" {
  ami                    = var.ec2_ami != "" ? var.ec2_ami : data.aws_ami.ubuntu.id
  instance_type          = var.ec2_instance_type
  key_name               = var.ec2_key_pair_name
  subnet_id              = aws_subnet.public[0].id
  vpc_security_group_ids = [aws_security_group.ec2.id]
  
  root_block_device {
    volume_type           = "gp3"
    volume_size           = var.ec2_root_volume_size
    delete_on_termination = false
    encrypted             = true
    
    tags = merge(var.tags, {
      Name = "${var.project_name}-${var.environment}-root-volume"
    })
  }

  user_data = templatefile("${path.module}/user_data.sh", {
    DOCKER_COMPOSE_VERSION = var.docker_compose_version
    AUTH_DB_HOST           = aws_db_instance.auth.address
    AUTH_DB_NAME           = var.auth_db_name
    USER_DB_HOST           = aws_db_instance.user.address
    USER_DB_NAME           = var.user_db_name
    PAYMENT_DB_HOST        = aws_db_instance.payment.address
    PAYMENT_DB_NAME        = var.payment_db_name
    ORDER_DB_HOST          = aws_db_instance.order.address
    ORDER_DB_NAME          = var.order_db_name
    DB_USERNAME            = var.db_username
    DB_PASSWORD            = var.db_password
  })

  user_data_replace_on_change = true

  tags = merge(var.tags, {
    Name = "${var.project_name}-${var.environment}-ec2"
  })

  lifecycle {
    ignore_changes = [user_data]
  }
}

# Associate Elastic IP with EC2
resource "aws_eip_association" "main" {
  instance_id   = aws_instance.main.id
  allocation_id = aws_eip.main.id
}

# CloudWatch Alarms for EC2
resource "aws_cloudwatch_metric_alarm" "ec2_cpu" {
  alarm_name          = "${var.project_name}-${var.environment}-ec2-cpu"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = "300"
  statistic           = "Average"
  threshold           = "80"
  alarm_description   = "This metric monitors ec2 cpu utilization"
  
  dimensions = {
    InstanceId = aws_instance.main.id
  }

  tags = var.tags
}

resource "aws_cloudwatch_metric_alarm" "ec2_status" {
  alarm_name          = "${var.project_name}-${var.environment}-ec2-status"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "StatusCheckFailed"
  namespace           = "AWS/EC2"
  period              = "60"
  statistic           = "Average"
  threshold           = "0"
  alarm_description   = "This metric monitors ec2 status checks"
  
  dimensions = {
    InstanceId = aws_instance.main.id
  }

  tags = var.tags
}
