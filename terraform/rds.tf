# Auth Database
resource "aws_db_instance" "auth" {
  identifier     = "${var.project_name}-${var.environment}-auth-db"
  engine         = "postgres"
  engine_version = var.rds_engine_version
  instance_class = var.rds_instance_class

  allocated_storage     = var.rds_allocated_storage
  max_allocated_storage = var.rds_max_allocated_storage
  storage_type          = "gp3"
  storage_encrypted     = true

  db_name  = var.auth_db_name
  username = var.db_username
  password = var.db_password

  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.rds.id]
  publicly_accessible    = false

  backup_retention_period = 7
  backup_window          = "03:00-04:00"
  maintenance_window     = "mon:04:00-mon:05:00"

  enabled_cloudwatch_logs_exports = ["postgresql", "upgrade"]
  
  skip_final_snapshot       = false
  final_snapshot_identifier = "${var.project_name}-${var.environment}-auth-db-final-snapshot"
  
  deletion_protection = true

  tags = merge(var.tags, {
    Name    = "${var.project_name}-${var.environment}-auth-db"
    Service = "auth-service"
  })
}

# User Database
resource "aws_db_instance" "user" {
  identifier     = "${var.project_name}-${var.environment}-user-db"
  engine         = "postgres"
  engine_version = var.rds_engine_version
  instance_class = var.rds_instance_class

  allocated_storage     = var.rds_allocated_storage
  max_allocated_storage = var.rds_max_allocated_storage
  storage_type          = "gp3"
  storage_encrypted     = true

  db_name  = var.user_db_name
  username = var.db_username
  password = var.db_password

  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.rds.id]
  publicly_accessible    = false

  backup_retention_period = 7
  backup_window          = "03:00-04:00"
  maintenance_window     = "mon:04:00-mon:05:00"

  enabled_cloudwatch_logs_exports = ["postgresql", "upgrade"]
  
  skip_final_snapshot       = false
  final_snapshot_identifier = "${var.project_name}-${var.environment}-user-db-final-snapshot"
  
  deletion_protection = true

  tags = merge(var.tags, {
    Name    = "${var.project_name}-${var.environment}-user-db"
    Service = "user-service"
  })
}

# Payment Database
resource "aws_db_instance" "payment" {
  identifier     = "${var.project_name}-${var.environment}-payment-db"
  engine         = "postgres"
  engine_version = var.rds_engine_version
  instance_class = var.rds_instance_class

  allocated_storage     = var.rds_allocated_storage
  max_allocated_storage = var.rds_max_allocated_storage
  storage_type          = "gp3"
  storage_encrypted     = true

  db_name  = var.payment_db_name
  username = var.db_username
  password = var.db_password

  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.rds.id]
  publicly_accessible    = false

  backup_retention_period = 7
  backup_window          = "03:00-04:00"
  maintenance_window     = "mon:04:00-mon:05:00"

  enabled_cloudwatch_logs_exports = ["postgresql", "upgrade"]
  
  skip_final_snapshot       = false
  final_snapshot_identifier = "${var.project_name}-${var.environment}-payment-db-final-snapshot"
  
  deletion_protection = true

  tags = merge(var.tags, {
    Name    = "${var.project_name}-${var.environment}-payment-db"
    Service = "payment-service"
  })
}

# Order Database
resource "aws_db_instance" "order" {
  identifier     = "${var.project_name}-${var.environment}-order-db"
  engine         = "postgres"
  engine_version = var.rds_engine_version
  instance_class = var.rds_instance_class

  allocated_storage     = var.rds_allocated_storage
  max_allocated_storage = var.rds_max_allocated_storage
  storage_type          = "gp3"
  storage_encrypted     = true

  db_name  = var.order_db_name
  username = var.db_username
  password = var.db_password

  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.rds.id]
  publicly_accessible    = false

  backup_retention_period = 7
  backup_window          = "03:00-04:00"
  maintenance_window     = "mon:04:00-mon:05:00"

  enabled_cloudwatch_logs_exports = ["postgresql", "upgrade"]
  
  skip_final_snapshot       = false
  final_snapshot_identifier = "${var.project_name}-${var.environment}-order-db-final-snapshot"
  
  deletion_protection = true

  tags = merge(var.tags, {
    Name    = "${var.project_name}-${var.environment}-order-db"
    Service = "order-service"
  })
}

# CloudWatch Alarms for RDS
resource "aws_cloudwatch_metric_alarm" "auth_db_cpu" {
  alarm_name          = "${var.project_name}-${var.environment}-auth-db-cpu"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "CPUUtilization"
  namespace           = "AWS/RDS"
  period              = "300"
  statistic           = "Average"
  threshold           = "80"
  
  dimensions = {
    DBInstanceIdentifier = aws_db_instance.auth.id
  }

  tags = var.tags
}

resource "aws_cloudwatch_metric_alarm" "auth_db_storage" {
  alarm_name          = "${var.project_name}-${var.environment}-auth-db-storage"
  comparison_operator = "LessThanThreshold"
  evaluation_periods  = "1"
  metric_name         = "FreeStorageSpace"
  namespace           = "AWS/RDS"
  period              = "300"
  statistic           = "Average"
  threshold           = "2000000000" # 2GB in bytes
  
  dimensions = {
    DBInstanceIdentifier = aws_db_instance.auth.id
  }

  tags = var.tags
}
