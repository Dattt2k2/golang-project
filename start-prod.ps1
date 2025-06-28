# Production Environment Script
Write-Host "Starting Production Environment..." -ForegroundColor Green

# Stop any running containers
Write-Host "Stopping existing containers..." -ForegroundColor Yellow
docker-compose down

# Build and start production environment
Write-Host "Building and starting production services..." -ForegroundColor Yellow
docker-compose up --build -d

Write-Host "Production environment started!" -ForegroundColor Green
Write-Host "Services available at:" -ForegroundColor Cyan
Write-Host "  - Kong Gateway: http://localhost:8000" -ForegroundColor White
Write-Host "  - Kong Admin: http://localhost:8001" -ForegroundColor White
Write-Host "  - Konga UI: http://localhost:1337" -ForegroundColor White
Write-Host "  - Auth Service: http://localhost:8081" -ForegroundColor White
Write-Host "  - Product Service: http://localhost:8082" -ForegroundColor White
Write-Host "  - Cart Service: http://localhost:8083" -ForegroundColor White
Write-Host "  - Order Service: http://localhost:8084" -ForegroundColor White
Write-Host "  - Search Service: http://localhost:8086" -ForegroundColor White
Write-Host "  - Mongo Express: http://localhost:8085" -ForegroundColor White
Write-Host "  - Elasticsearch: http://localhost:9200" -ForegroundColor White
Write-Host "  - Kibana: http://localhost:5601" -ForegroundColor White
Write-Host ""
Write-Host "To view logs: docker-compose logs -f [service-name]" -ForegroundColor Yellow
