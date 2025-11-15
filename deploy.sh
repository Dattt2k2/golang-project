#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}============================================${NC}"
echo -e "${GREEN}  Microservices Deployment Script${NC}"
echo -e "${GREEN}============================================${NC}"

# Check if running as root or with sudo
if [ "$EUID" -eq 0 ]; then 
    DOCKER_CMD="docker"
    COMPOSE_CMD="docker-compose"
else
    DOCKER_CMD="sudo docker"
    COMPOSE_CMD="sudo docker-compose"
fi

# Check if .env file exists
if [ ! -f .env ]; then
    echo -e "${RED}Error: .env file not found!${NC}"
    echo "Please create .env file with required environment variables"
    exit 1
fi

# Load environment variables
export $(cat .env | grep -v '^#' | xargs)

# Check if docker-compose.prod.yaml exists
if [ ! -f docker-compose.prod.yaml ]; then
    echo -e "${RED}Error: docker-compose.prod.yaml not found!${NC}"
    exit 1
fi

# Function to check if a service is healthy
check_service_health() {
    local service=$1
    local max_attempts=30
    local attempt=1
    
    echo -e "${YELLOW}Checking $service health...${NC}"
    
    while [ $attempt -le $max_attempts ]; do
        if $COMPOSE_CMD -f docker-compose.prod.yaml ps $service | grep -q "healthy\|running"; then
            echo -e "${GREEN}$service is healthy${NC}"
            return 0
        fi
        echo "Waiting for $service... (attempt $attempt/$max_attempts)"
        sleep 5
        ((attempt++))
    done
    
    echo -e "${RED}$service failed to become healthy${NC}"
    return 1
}

# Backup database before deployment (if needed)
backup_database() {
    echo -e "${YELLOW}Creating database backup...${NC}"
    BACKUP_DIR="./backups"
    mkdir -p $BACKUP_DIR
    TIMESTAMP=$(date +%Y%m%d_%H%M%S)
    
    # Add your backup commands here if needed
    # Example: pg_dump -h $DB_HOST -U $DB_USER $DB_NAME > $BACKUP_DIR/backup_$TIMESTAMP.sql
    
    echo -e "${GREEN}Backup created at $BACKUP_DIR/backup_$TIMESTAMP${NC}"
}

# Main deployment process
main() {
    echo -e "${YELLOW}Starting deployment process...${NC}"
    
    # Optional: Create backup
    # backup_database
    
    # Pull latest images
    echo -e "${YELLOW}Pulling latest Docker images...${NC}"
    $COMPOSE_CMD -f docker-compose.prod.yaml pull
    
    # Stop existing services
    echo -e "${YELLOW}Stopping existing services...${NC}"
    $COMPOSE_CMD -f docker-compose.prod.yaml down
    
    # Remove unused images and volumes (optional)
    echo -e "${YELLOW}Cleaning up unused Docker resources...${NC}"
    $DOCKER_CMD system prune -f
    
    # Start infrastructure services first
    echo -e "${YELLOW}Starting infrastructure services...${NC}"
    $COMPOSE_CMD -f docker-compose.prod.yaml up -d redis
    sleep 5
    
    $COMPOSE_CMD -f docker-compose.prod.yaml up -d kafka elasticsearch
    sleep 10
    check_service_health kafka
    check_service_health elasticsearch
    
    # Start application services
    echo -e "${YELLOW}Starting application services...${NC}"
    $COMPOSE_CMD -f docker-compose.prod.yaml up -d \
        auth-service \
        user-service \
        product-service \
        cart-service \
        order-service \
        payment-service \
        search-service \
        email-service \
        review-service
    
    # Wait for services to be healthy
    sleep 15
    
    # Start API Gateway and supporting services
    echo -e "${YELLOW}Starting API Gateway and supporting services...${NC}"
    $COMPOSE_CMD -f docker-compose.prod.yaml up -d \
        api-gateway \
        traefik \
        kibana
    
    # Wait for all services to stabilize
    echo -e "${YELLOW}Waiting for all services to stabilize...${NC}"
    sleep 20
    
    # Check service status
    echo -e "${YELLOW}Service Status:${NC}"
    $COMPOSE_CMD -f docker-compose.prod.yaml ps
    
    # Show logs
    echo -e "${YELLOW}Recent logs:${NC}"
    $COMPOSE_CMD -f docker-compose.prod.yaml logs --tail=50
    
    # Health check
    echo -e "${YELLOW}Running health checks...${NC}"
    FAILED_SERVICES=()
    
    for service in api-gateway auth-service user-service product-service order-service payment-service; do
        if ! check_service_health $service; then
            FAILED_SERVICES+=($service)
        fi
    done
    
    # Final status
    echo -e "${GREEN}============================================${NC}"
    if [ ${#FAILED_SERVICES[@]} -eq 0 ]; then
        echo -e "${GREEN}✓ Deployment completed successfully!${NC}"
        echo -e "${GREEN}============================================${NC}"
        echo ""
        echo "Services are running at:"
        echo "- API Gateway: http://$(curl -s ifconfig.me):8080"
        echo "- Traefik Dashboard: http://$(curl -s ifconfig.me):8081"
        echo "- Kibana: http://$(curl -s ifconfig.me):5601"
        echo ""
        echo "To view logs: $COMPOSE_CMD -f docker-compose.prod.yaml logs -f [service]"
        echo "To check status: $COMPOSE_CMD -f docker-compose.prod.yaml ps"
    else
        echo -e "${RED}✗ Deployment completed with errors${NC}"
        echo -e "${RED}Failed services: ${FAILED_SERVICES[@]}${NC}"
        echo -e "${RED}============================================${NC}"
        echo ""
        echo "Check logs for failed services:"
        for service in "${FAILED_SERVICES[@]}"; do
            echo "  $COMPOSE_CMD -f docker-compose.prod.yaml logs $service"
        done
        exit 1
    fi
}

# Run main deployment
main
