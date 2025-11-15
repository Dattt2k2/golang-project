#!/bin/bash
set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}============================================${NC}"
echo -e "${GREEN}  Database Backup Script${NC}"
echo -e "${GREEN}============================================${NC}"

# Configuration
BACKUP_DIR="./backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS=7

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
else
    echo -e "${RED}Error: .env file not found!${NC}"
    exit 1
fi

# Create backup directory
mkdir -p $BACKUP_DIR

# Function to backup a database
backup_database() {
    local db_host=$1
    local db_name=$2
    local db_user=$3
    local db_password=$4
    local service_name=$5
    
    echo -e "${YELLOW}Backing up $service_name database...${NC}"
    
    BACKUP_FILE="$BACKUP_DIR/${service_name}_${TIMESTAMP}.sql"
    
    # Use Docker to run pg_dump
    PGPASSWORD=$db_password docker run --rm \
        --network host \
        postgres:15 \
        pg_dump -h $db_host -U $db_user -d $db_name \
        > $BACKUP_FILE
    
    if [ $? -eq 0 ]; then
        # Compress backup
        gzip $BACKUP_FILE
        echo -e "${GREEN}✓ Backup completed: ${BACKUP_FILE}.gz${NC}"
        
        # Calculate size
        SIZE=$(du -h "${BACKUP_FILE}.gz" | cut -f1)
        echo -e "${GREEN}  Size: $SIZE${NC}"
    else
        echo -e "${RED}✗ Backup failed for $service_name${NC}"
        return 1
    fi
}

# Backup all databases
echo -e "${YELLOW}Starting backup process...${NC}"

backup_database "$AUTH_DB_HOST" "$AUTH_DB_NAME" "$DB_USERNAME" "$DB_PASSWORD" "auth"
backup_database "$USER_DB_HOST" "$USER_DB_NAME" "$DB_USERNAME" "$DB_PASSWORD" "user"
backup_database "$PAYMENT_DB_HOST" "$PAYMENT_DB_NAME" "$DB_USERNAME" "$DB_PASSWORD" "payment"
backup_database "$ORDER_DB_HOST" "$ORDER_DB_NAME" "$DB_USERNAME" "$DB_PASSWORD" "order"

# Clean old backups
echo -e "${YELLOW}Cleaning old backups (older than $RETENTION_DAYS days)...${NC}"
find $BACKUP_DIR -name "*.sql.gz" -mtime +$RETENTION_DAYS -delete
echo -e "${GREEN}✓ Old backups cleaned${NC}"

# List all backups
echo -e "${YELLOW}Current backups:${NC}"
ls -lh $BACKUP_DIR/*.sql.gz 2>/dev/null || echo "No backups found"

# Optional: Upload to S3
if [ ! -z "$AWS_S3_BACKUP_BUCKET" ]; then
    echo -e "${YELLOW}Uploading backups to S3...${NC}"
    aws s3 sync $BACKUP_DIR s3://$AWS_S3_BACKUP_BUCKET/database-backups/ \
        --exclude "*" --include "*.sql.gz"
    echo -e "${GREEN}✓ Backups uploaded to S3${NC}"
fi

echo -e "${GREEN}============================================${NC}"
echo -e "${GREEN}Backup process completed!${NC}"
echo -e "${GREEN}============================================${NC}"
