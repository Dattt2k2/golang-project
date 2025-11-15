# Application Load Balancer
resource "aws_lb" "main" {
  name               = "${var.project_name}-${var.environment}-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = aws_subnet.public[*].id

  enable_deletion_protection = true
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

resource "aws_lb_target_group" "api_gateway" {
  name     = "${var.project_name}-${var.environment}-gateway-tg"
  port     = 8080
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
    Name    = "${var.project_name}-${var.environment}-gateway-tg"
    Service = "api-gateway"
  })
}

# ALB Listener
resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.main.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.api_gateway.arn
  }

  tags = var.tags
}

# ALB Listener Rules for path-based routing
resource "aws_lb_listener_rule" "auth" {
  listener_arn = aws_lb_listener.http.arn
  priority     = 100

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.auth_service.arn
  }

  condition {
    path_pattern {
      values = ["/api/auth/*", "/auth/*"]
    }
  }
}

resource "aws_lb_listener_rule" "user" {
  listener_arn = aws_lb_listener.http.arn
  priority     = 101

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.user_service.arn
  }

  condition {
    path_pattern {
      values = ["/api/user/*", "/user/*"]
    }
  }
}

resource "aws_lb_listener_rule" "product" {
  listener_arn = aws_lb_listener.http.arn
  priority     = 102

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.product_service.arn
  }

  condition {
    path_pattern {
      values = ["/api/product/*", "/product/*"]
    }
  }
}

resource "aws_lb_listener_rule" "order" {
  listener_arn = aws_lb_listener.http.arn
  priority     = 103

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.order_service.arn
  }

  condition {
    path_pattern {
      values = ["/api/order/*", "/order/*"]
    }
  }
}

resource "aws_lb_listener_rule" "payment" {
  listener_arn = aws_lb_listener.http.arn
  priority     = 104

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.payment_service.arn
  }

  condition {
    path_pattern {
      values = ["/api/payment/*", "/payment/*"]
    }
  }
}

# EC2 Instances for each service
locals {
  services = {
    api-gateway = {
      port          = 8080
      target_group  = aws_lb_target_group.api_gateway.arn
      instance_type = var.gateway_instance_type
    }
    auth-service = {
      port          = 8081
      target_group  = aws_lb_target_group.auth_service.arn
      instance_type = var.service_instance_type
    }
    user-service = {
      port          = 8082
      target_group  = aws_lb_target_group.user_service.arn
      instance_type = var.service_instance_type
    }
    product-service = {
      port          = 8083
      target_group  = aws_lb_target_group.product_service.arn
      instance_type = var.service_instance_type
    }
    order-service = {
      port          = 8084
      target_group  = aws_lb_target_group.order_service.arn
      instance_type = var.service_instance_type
    }
    payment-service = {
      port          = 8086
      target_group  = aws_lb_target_group.payment_service.arn
      instance_type = var.service_instance_type
    }
  }
}

# EC2 Instances
resource "aws_instance" "services" {
  for_each = local.services

  ami                    = var.ec2_ami != "" ? var.ec2_ami : data.aws_ami.ubuntu.id
  instance_type          = each.value.instance_type
  key_name               = var.ec2_key_pair_name
  subnet_id              = aws_subnet.public[0].id
  vpc_security_group_ids = [aws_security_group.ec2_services.id]

  root_block_device {
    volume_type           = "gp3"
    volume_size           = 30
    delete_on_termination = false
    encrypted             = true
  }

  user_data = templatefile("${path.module}/service_user_data.sh", {
    SERVICE_NAME   = each.key
    SERVICE_PORT   = each.value.port
    AUTH_DB_HOST   = aws_db_instance.auth.address
    AUTH_DB_NAME   = var.auth_db_name
    USER_DB_HOST   = aws_db_instance.user.address
    USER_DB_NAME   = var.user_db_name
    PAYMENT_DB_HOST = aws_db_instance.payment.address
    PAYMENT_DB_NAME = var.payment_db_name
    ORDER_DB_HOST   = aws_db_instance.order.address
    ORDER_DB_NAME   = var.order_db_name
    DB_USERNAME     = var.db_username
    DB_PASSWORD     = var.db_password
    REDIS_HOST      = aws_instance.infrastructure["redis"].private_ip
    KAFKA_HOST      = aws_instance.infrastructure["kafka"].private_ip
    ELASTICSEARCH_HOST = aws_instance.infrastructure["elasticsearch"].private_ip
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

# Target Group Attachments
resource "aws_lb_target_group_attachment" "services" {
  for_each = local.services

  target_group_arn = each.value.target_group
  target_id        = aws_instance.services[each.key].id
  port             = each.value.port
}

# Infrastructure services (Redis, Kafka, Elasticsearch) on dedicated instances
locals {
  infrastructure = {
    redis = {
      instance_type = var.infrastructure_instance_type
      port         = 6379
    }
    kafka = {
      instance_type = var.infrastructure_instance_type
      port         = 9092
    }
    elasticsearch = {
      instance_type = var.infrastructure_instance_type
      port         = 9200
    }
  }
}

resource "aws_instance" "infrastructure" {
  for_each = local.infrastructure

  ami                    = var.ec2_ami != "" ? var.ec2_ami : data.aws_ami.ubuntu.id
  instance_type          = each.value.instance_type
  key_name               = var.ec2_key_pair_name
  subnet_id              = aws_subnet.private[0].id
  vpc_security_group_ids = [aws_security_group.ec2_infrastructure.id]

  root_block_device {
    volume_type           = "gp3"
    volume_size           = 50
    delete_on_termination = false
    encrypted             = true
  }

  user_data = templatefile("${path.module}/infrastructure_user_data.sh", {
    SERVICE_NAME = each.key
    SERVICE_PORT = each.value.port
  })

  user_data_replace_on_change = true

  tags = merge(var.tags, {
    Name = "${var.project_name}-${var.environment}-${each.key}"
    Type = "Infrastructure"
  })

  lifecycle {
    ignore_changes = [user_data]
  }
}

# CloudWatch Alarms for each service
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

resource "aws_cloudwatch_metric_alarm" "service_status" {
  for_each = local.services

  alarm_name          = "${var.project_name}-${var.environment}-${each.key}-status"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "StatusCheckFailed"
  namespace           = "AWS/EC2"
  period              = "60"
  statistic           = "Average"
  threshold           = "0"
  alarm_description   = "Status check for ${each.key}"

  dimensions = {
    InstanceId = aws_instance.services[each.key].id
  }

  tags = var.tags
}

# CloudWatch Alarms for infrastructure
resource "aws_cloudwatch_metric_alarm" "infrastructure_cpu" {
  for_each = local.infrastructure

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
    InstanceId = aws_instance.infrastructure[each.key].id
  }

  tags = var.tags
}
