#!/bin/bash
set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}============================================${NC}"
echo -e "${GREEN}  Database Restore Script${NC}"
echo -e "${GREEN}============================================${NC}"

# Check arguments
if [ $# -lt 2 ]; then
    echo -e "${RED}Usage: $0 <service_name> <backup_file>${NC}"
    echo ""
    echo "Available services: auth, user, payment, order"
    echo "Example: $0 auth ./backups/auth_20240101_120000.sql.gz"
    echo ""
    echo "Available backups:"
    ls -lh ./backups/*.sql.gz 2>/dev/null || echo "No backups found"
    exit 1
fi

SERVICE_NAME=$1
BACKUP_FILE=$2

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
else
    echo -e "${RED}Error: .env file not found!${NC}"
    exit 1
fi

# Check if backup file exists
if [ ! -f "$BACKUP_FILE" ]; then
    echo -e "${RED}Error: Backup file not found: $BACKUP_FILE${NC}"
    exit 1
fi

# Map service name to database variables
case $SERVICE_NAME in
    auth)
        DB_HOST=$AUTH_DB_HOST
        DB_NAME=$AUTH_DB_NAME
        ;;
    user)
        DB_HOST=$USER_DB_HOST
        DB_NAME=$USER_DB_NAME
        ;;
    payment)
        DB_HOST=$PAYMENT_DB_HOST
        DB_NAME=$PAYMENT_DB_NAME
        ;;
    order)
        DB_HOST=$ORDER_DB_HOST
        DB_NAME=$ORDER_DB_NAME
        ;;
    *)
        echo -e "${RED}Error: Invalid service name: $SERVICE_NAME${NC}"
        echo "Valid options: auth, user, payment, order"
        exit 1
        ;;
esac

# Confirmation
echo -e "${YELLOW}WARNING: This will restore $SERVICE_NAME database from backup${NC}"
echo -e "${YELLOW}Database: $DB_NAME${NC}"
echo -e "${YELLOW}Host: $DB_HOST${NC}"
echo -e "${YELLOW}Backup: $BACKUP_FILE${NC}"
echo ""
read -p "Are you sure you want to continue? (yes/no): " CONFIRM

if [ "$CONFIRM" != "yes" ]; then
    echo -e "${RED}Restore cancelled${NC}"
    exit 0
fi

# Create a backup of current database before restore
echo -e "${YELLOW}Creating safety backup of current database...${NC}"
SAFETY_BACKUP="./backups/${SERVICE_NAME}_before_restore_$(date +%Y%m%d_%H%M%S).sql"
PGPASSWORD=$DB_PASSWORD docker run --rm \
    --network host \
    postgres:15 \
    pg_dump -h $DB_HOST -U $DB_USERNAME -d $DB_NAME \
    > $SAFETY_BACKUP

if [ $? -eq 0 ]; then
    gzip $SAFETY_BACKUP
    echo -e "${GREEN}✓ Safety backup created: ${SAFETY_BACKUP}.gz${NC}"
else
    echo -e "${RED}✗ Failed to create safety backup${NC}"
    exit 1
fi

# Decompress backup if needed
RESTORE_FILE=$BACKUP_FILE
if [[ $BACKUP_FILE == *.gz ]]; then
    echo -e "${YELLOW}Decompressing backup...${NC}"
    RESTORE_FILE="${BACKUP_FILE%.gz}"
    gunzip -c $BACKUP_FILE > $RESTORE_FILE
fi

# Restore database
echo -e "${YELLOW}Restoring database...${NC}"
PGPASSWORD=$DB_PASSWORD docker run --rm -i \
    --network host \
    postgres:15 \
    psql -h $DB_HOST -U $DB_USERNAME -d $DB_NAME \
    < $RESTORE_FILE

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Database restored successfully${NC}"
    
    # Clean up decompressed file if it was created
    if [[ $BACKUP_FILE == *.gz ]]; then
        rm $RESTORE_FILE
    fi
else
    echo -e "${RED}✗ Database restore failed${NC}"
    echo -e "${YELLOW}Safety backup is available at: ${SAFETY_BACKUP}.gz${NC}"
    exit 1
fi

# Restart affected service
echo -e "${YELLOW}Restarting ${SERVICE_NAME}-service...${NC}"
if [ -f docker-compose.prod.yaml ]; then
    sudo docker-compose -f docker-compose.prod.yaml restart ${SERVICE_NAME}-service
    echo -e "${GREEN}✓ Service restarted${NC}"
fi

echo -e "${GREEN}============================================${NC}"
echo -e "${GREEN}Restore completed successfully!${NC}"
echo -e "${GREEN}============================================${NC}"
