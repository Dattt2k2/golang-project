# Environment Variables Validation Script
Write-Host "Validating all .env files..." -ForegroundColor Yellow
Write-Host ""

$services = @("api-gateway", "auth-service", "product-service", "cart-service", "order-service", "search-service")
$allValid = $true

foreach ($service in $services) {
    $envFile = "./$service/.env"
    Write-Host "Checking $service..." -ForegroundColor Cyan
    
    if (Test-Path $envFile) {
        $content = Get-Content $envFile
        $hasSecrets = $false
        
        # Check for common secret patterns
        foreach ($line in $content) {
            if ($line -match "(SECRET|KEY|PASSWORD|TOKEN)" -and $line -notmatch "^#") {
                $key = $line.Split('=')[0]
                $value = $line.Split('=')[1]
                
                if ($value -eq "" -or $value -eq "your-" -or $value -match "your-.*-here") {
                    Write-Host "  ⚠️  $key needs to be configured" -ForegroundColor Red
                    $allValid = $false
                } else {
                    Write-Host "  ✅ $key is configured" -ForegroundColor Green
                    $hasSecrets = $true
                }
            }
        }
        
        if (-not $hasSecrets) {
            Write-Host "  ℹ️  No secrets found in this service" -ForegroundColor Gray
        }
    } else {
        Write-Host "  ❌ .env file not found!" -ForegroundColor Red
        $allValid = $false
    }
    
    Write-Host ""
}

if ($allValid) {
    Write-Host "✅ All environment files are properly configured!" -ForegroundColor Green
    Write-Host "You can safely run: .\start-dev.ps1" -ForegroundColor Yellow
} else {
    Write-Host "❌ Some environment variables need configuration." -ForegroundColor Red
    Write-Host "Please update the marked variables before starting services." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "💡 Tips:" -ForegroundColor Cyan
Write-Host "  - Never commit .env files with real secrets" -ForegroundColor White
Write-Host "  - Use different secrets for production" -ForegroundColor White
Write-Host "  - Keep AWS/DB credentials secure" -ForegroundColor White
