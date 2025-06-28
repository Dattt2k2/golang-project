# Development Environment Script
Write-Host "Starting Development Environment with Hot Reload..." -ForegroundColor Green

# Stop any running containers
Write-Host "Stopping existing containers..." -ForegroundColor Yellow
docker-compose -f docker-compose.dev.yaml down

# Build and start development environment
Write-Host "Building and starting development services..." -ForegroundColor Yellow
docker-compose -f docker-compose.dev.yaml up --build -d

Write-Host "Development environment started!" -ForegroundColor Green
Write-Host "Services available at:" -ForegroundColor Cyan
Write-Host "  - API Gateway (Main Entry): http://localhost:8080" -ForegroundColor Yellow
Write-Host "  - Auth Service: http://localhost:8081" -ForegroundColor White
Write-Host "  - Product Service: http://localhost:8082" -ForegroundColor White
Write-Host "  - Cart Service: http://localhost:8083" -ForegroundColor White
Write-Host "  - Order Service: http://localhost:8084" -ForegroundColor White
Write-Host "  - Search Service: http://localhost:8086" -ForegroundColor White
Write-Host "  - Mongo Express: http://localhost:8085" -ForegroundColor White
Write-Host "  - Elasticsearch: http://localhost:9200" -ForegroundColor White
Write-Host ""
Write-Host "API Gateway routes all requests to appropriate microservices" -ForegroundColor Cyan
Write-Host "Example: http://localhost:8080/auth/users/register" -ForegroundColor Yellow
Write-Host "Hot reload is enabled - code changes will automatically restart services!" -ForegroundColor Green
Write-Host "To view logs: docker-compose -f docker-compose.dev.yaml logs -f [service-name]" -ForegroundColor Yellow
