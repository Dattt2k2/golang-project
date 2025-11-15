# Application Load Balancer
resource "aws_lb" "main" {
  name               = "${var.project_name}-${var.environment}-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = aws_subnet.public[*].id

  enable_deletion_protection = false
  enable_http2              = true

  tags = merge(var.tags, {
    Name = "${var.project_name}-${var.environment}-alb"
  })
}

# Target Groups for each service
resource "aws_lb_target_group" "auth_service" {
  name     = "${var.project_name}-${var.environment}-auth-tg"
  port     = 8081
  protocol = "HTTP"
  vpc_id   = aws_vpc.main.id

  health_check {
    enabled             = true
    healthy_threshold   = 2
    interval            = 30
    matcher             = "200"
    path                = "/health"
    port                = "traffic-port"
    protocol            = "HTTP"
    timeout             = 5
    unhealthy_threshold = 2
  }

  tags = merge(var.tags, {
    Name    = "${var.project_name}-${var.environment}-auth-tg"
    Service = "auth-service"
  })
}

resource "aws_lb_target_group" "user_service" {
  name     = "${var.project_name}-${var.environment}-user-tg"
  port     = 8082
  protocol = "HTTP"
  vpc_id   = aws_vpc.main.id

  health_check {
    enabled             = true
    healthy_threshold   = 2
    interval            = 30
    matcher             = "200"
    path                = "/health"
    port                = "traffic-port"
    protocol            = "HTTP"
    timeout             = 5
    unhealthy_threshold = 2
  }

  tags = merge(var.tags, {
    Name    = "${var.project_name}-${var.environment}-user-tg"
    Service = "user-service"
  })
}

resource "aws_lb_target_group" "product_service" {
  name     = "${var.project_name}-${var.environment}-product-tg"
  port     = 8083
  protocol = "HTTP"
  vpc_id   = aws_vpc.main.id

  health_check {
    enabled             = true
    healthy_threshold   = 2
    interval            = 30
    matcher             = "200"
    path                = "/health"
    port                = "traffic-port"
    protocol            = "HTTP"
    timeout             = 5
    unhealthy_threshold = 2
  }

  tags = merge(var.tags, {
    Name    = "${var.project_name}-${var.environment}-product-tg"
    Service = "product-service"
  })
}

resource "aws_lb_target_group" "order_service" {
  name     = "${var.project_name}-${var.environment}-order-tg"
  port     = 8084
  protocol = "HTTP"
  vpc_id   = aws_vpc.main.id

  health_check {
    enabled             = true
    healthy_threshold   = 2
    interval            = 30
    matcher             = "200"
    path                = "/health"
    port                = "traffic-port"
    protocol            = "HTTP"
    timeout             = 5
    unhealthy_threshold = 2
  }

  tags = merge(var.tags, {
    Name    = "${var.project_name}-${var.environment}-order-tg"
    Service = "order-service"
  })
}

resource "aws_lb_target_group" "payment_service" {
  name     = "${var.project_name}-${var.environment}-payment-tg"
  port     = 8086
  protocol = "HTTP"
  vpc_id   = aws_vpc.main.id

  health_check {
    enabled             = true
    healthy_threshold   = 2
    interval            = 30
    matcher             = "200"
    path                = "/health"
    port                = "traffic-port"
    protocol            = "HTTP"
    timeout             = 5
    unhealthy_threshold = 2
  }

  tags = merge(var.tags, {
    Name    = "${var.project_name}-${var.environment}-payment-tg"
    Service = "payment-service"
  })
}

resource "aws_lb_target_group" "cart_service" {
  name     = "${var.project_name}-${var.environment}-cart-tg"
  port     = 8085
  protocol = "HTTP"
  vpc_id   = aws_vpc.main.id

  health_check {
    enabled             = true
    healthy_threshold   = 2
    interval            = 30
    matcher             = "200"
    path                = "/health"
    port                = "traffic-port"
    protocol            = "HTTP"
    timeout             = 5
    unhealthy_threshold = 2
  }

  tags = merge(var.tags, {
    Name    = "${var.project_name}-${var.environment}-cart-tg"
    Service = "cart-service"
  })
}

resource "aws_lb_target_group" "traefik" {
  name     = "${var.project_name}-${var.environment}-traefik-tg"
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.main.id

  health_check {
    enabled             = true
    healthy_threshold   = 2
    interval            = 30
    matcher             = "200"
    path                = "/ping"
    port                = "traffic-port"
    protocol            = "HTTP"
    timeout             = 5
    unhealthy_threshold = 2
  }

  tags = merge(var.tags, {
    Name    = "${var.project_name}-${var.environment}-traefik-tg"
    Service = "traefik"
  })
}

# ALB Listener - Forward all to Traefik
resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.main.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.traefik.arn
  }

  tags = var.tags
}

# Service instances configuration
locals {
  services = {
    auth-service = {
      port          = 8081
      target_group  = aws_lb_target_group.auth_service.arn
      instance_type = var.service_instance_type
      db_host       = aws_db_instance.auth.address
      db_name       = var.auth_db_name
    }
    user-service = {
      port          = 8082
      target_group  = aws_lb_target_group.user_service.arn
      instance_type = var.service_instance_type
      db_host       = aws_db_instance.user.address
      db_name       = var.user_db_name
    }
    product-service = {
      port          = 8083
      target_group  = aws_lb_target_group.product_service.arn
      instance_type = var.service_instance_type
      db_host       = aws_db_instance.auth.address
      db_name       = "productdb"
    }
    cart-service = {
      port          = 8085
      target_group  = aws_lb_target_group.cart_service.arn
      instance_type = var.service_instance_type
      db_host       = aws_db_instance.auth.address
      db_name       = "cartdb"
    }
    order-service = {
      port          = 8084
      target_group  = aws_lb_target_group.order_service.arn
      instance_type = var.service_instance_type
      db_host       = aws_db_instance.order.address
      db_name       = var.order_db_name
    }
    payment-service = {
      port          = 8086
      target_group  = aws_lb_target_group.payment_service.arn
      instance_type = var.service_instance_type
      db_host       = aws_db_instance.payment.address
      db_name       = var.payment_db_name
    }
  }
}

# EC2 Instances for microservices
resource "aws_instance" "services" {
  for_each = local.services

  ami                    = var.ec2_ami != "" ? var.ec2_ami : data.aws_ami.ubuntu.id
  instance_type          = each.value.instance_type
  key_name               = var.ec2_key_pair_name
  subnet_id              = aws_subnet.public[0].id
  vpc_security_group_ids = [aws_security_group.ec2_services.id]

  root_block_device {
    volume_type           = "gp3"
    volume_size           = 20
    delete_on_termination = false
    encrypted             = true
  }

  user_data = templatefile("${path.module}/service_user_data.sh", {
    SERVICE_NAME        = each.key
    SERVICE_PORT        = each.value.port
    DB_HOST            = each.value.db_host
    DB_NAME            = each.value.db_name
    DB_USERNAME        = var.db_username
    DB_PASSWORD        = var.db_password
    REDIS_HOST         = aws_instance.shared_infrastructure.private_ip
    KAFKA_HOST         = aws_instance.shared_infrastructure.private_ip
    ELASTICSEARCH_HOST = aws_instance.shared_infrastructure.private_ip
  })

  user_data_replace_on_change = true

  tags = merge(var.tags, {
    Name    = "${var.project_name}-${var.environment}-${each.key}"
    Service = each.key
  })

  lifecycle {
    ignore_changes = [user_data]
  }
}

# Target Group Attachments for services
resource "aws_lb_target_group_attachment" "services" {
  for_each = local.services

  target_group_arn = each.value.target_group
  target_id        = aws_instance.services[each.key].id
  port             = each.value.port
}

# Shared Infrastructure Instance (Redis, Kafka, Elasticsearch)
resource "aws_instance" "shared_infrastructure" {
  ami                    = var.ec2_ami != "" ? var.ec2_ami : data.aws_ami.ubuntu.id
  instance_type          = var.shared_infrastructure_instance_type
  key_name               = var.ec2_key_pair_name
  subnet_id              = aws_subnet.private[0].id
  vpc_security_group_ids = [aws_security_group.ec2_infrastructure.id]

  root_block_device {
    volume_type           = "gp3"
    volume_size           = 100
    delete_on_termination = false
    encrypted             = true
  }

  user_data = templatefile("${path.module}/shared_infrastructure_user_data.sh", {
    ENVIRONMENT = var.environment
  })

  user_data_replace_on_change = true

  tags = merge(var.tags, {
    Name = "${var.project_name}-${var.environment}-shared-infrastructure"
    Type = "Infrastructure"
  })

  lifecycle {
    ignore_changes = [user_data]
  }
}

# Traefik Instance (API Gateway)
resource "aws_instance" "traefik" {
  ami                    = var.ec2_ami != "" ? var.ec2_ami : data.aws_ami.ubuntu.id
  instance_type          = var.gateway_instance_type
  key_name               = var.ec2_key_pair_name
  subnet_id              = aws_subnet.public[0].id
  vpc_security_group_ids = [aws_security_group.ec2_traefik.id]

  root_block_device {
    volume_type           = "gp3"
    volume_size           = 20
    delete_on_termination = false
    encrypted             = true
  }

  user_data = templatefile("${path.module}/traefik_user_data.sh", {
    AUTH_SERVICE_IP    = aws_instance.services["auth-service"].private_ip
    USER_SERVICE_IP    = aws_instance.services["user-service"].private_ip
    PRODUCT_SERVICE_IP = aws_instance.services["product-service"].private_ip
    CART_SERVICE_IP    = aws_instance.services["cart-service"].private_ip
    ORDER_SERVICE_IP   = aws_instance.services["order-service"].private_ip
    PAYMENT_SERVICE_IP = aws_instance.services["payment-service"].private_ip
  })

  user_data_replace_on_change = true

  tags = merge(var.tags, {
    Name    = "${var.project_name}-${var.environment}-traefik"
    Service = "api-gateway"
  })

  lifecycle {
    ignore_changes = [user_data]
  }
}

# Target Group Attachment for Traefik
resource "aws_lb_target_group_attachment" "traefik" {
  target_group_arn = aws_lb_target_group.traefik.arn
  target_id        = aws_instance.traefik.id
  port             = 80
}

# CloudWatch Alarms for services
resource "aws_cloudwatch_metric_alarm" "service_cpu" {
  for_each = local.services

  alarm_name          = "${var.project_name}-${var.environment}-${each.key}-cpu"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = "300"
  statistic           = "Average"
  threshold           = "80"
  alarm_description   = "CPU utilization for ${each.key}"

  dimensions = {
    InstanceId = aws_instance.services[each.key].id
  }

  tags = var.tags
}

# CloudWatch Alarm for shared infrastructure
resource "aws_cloudwatch_metric_alarm" "infrastructure_cpu" {
  alarm_name          = "${var.project_name}-${var.environment}-shared-infra-cpu"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = "300"
  statistic           = "Average"
  threshold           = "80"
  alarm_description   = "CPU utilization for shared infrastructure"

  dimensions = {
    InstanceId = aws_instance.shared_infrastructure.id
  }

  tags = var.tags
}

# CloudWatch Alarm for Traefik
resource "aws_cloudwatch_metric_alarm" "traefik_cpu" {
  alarm_name          = "${var.project_name}-${var.environment}-traefik-cpu"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = "300"
  statistic           = "Average"
  threshold           = "80"
  alarm_description   = "CPU utilization for Traefik"

  dimensions = {
    InstanceId = aws_instance.traefik.id
  }

  tags = var.tags
}
