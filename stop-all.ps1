# Stop All Services Script
Write-Host "Stopping all Docker services..." -ForegroundColor Yellow

# Stop development environment
Write-Host "Stopping development environment..." -ForegroundColor Cyan
docker-compose -f docker-compose.dev.yaml down

# Stop production environment
Write-Host "Stopping production environment..." -ForegroundColor Cyan
docker-compose down

# Clean up unused containers and networks
Write-Host "Cleaning up unused containers and networks..." -ForegroundColor Cyan
docker system prune -f

Write-Host "All services stopped and cleaned up!" -ForegroundColor Green
