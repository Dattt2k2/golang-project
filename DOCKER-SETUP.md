# Docker Development Setup

## 🚀 Quick Start

### Development Environment (Hot Reload)
```powershell
.\start-dev.ps1
```
- Sử dụng `Dockerfile.dev` với Air hot reload + Go 1.23
- Code thay đổi sẽ tự động restart service
- Sử dụng API Gateway tự custom trên port 8080 (không có Kong)
- Tất cả request đi qua: `http://localhost:8080/[service-prefix]/[endpoint]`

### Production Environment
```powershell
.\start-prod.ps1
```
- Sử dụng `Dockerfile` production build
- Có đầy đủ Kong Gateway, Kibana, Filebeat
- Tối ưu cho performance

### Stop All Services
```powershell
.\stop-all.ps1
```

## 🛠️ Development Features

### Hot Reload với Air
- **Air** tự động phát hiện thay đổi file `.go`
- Rebuild và restart chỉ trong vài giây
- Log real-time để debug

### Volume Mapping
- Source code được mount vào container
- Thay đổi code ngay lập tức có hiệu lực
- Không cần rebuild image

## 📊 Service Ports

### Development Environment
- **API Gateway (Main Entry)**: http://localhost:8080
- **Auth Service**: http://localhost:8081  
- **Product Service**: http://localhost:8082
- **Cart Service**: http://localhost:8083
- **Order Service**: http://localhost:8084
- **Search Service**: http://localhost:8086
- **Mongo Express**: http://localhost:8085

### Example API calls:
- Register: `POST http://localhost:8080/auth/users/register`
- Login: `POST http://localhost:8080/auth/users/login`
- Products: `GET http://localhost:8080/products`

### Production Environment
- **Kong Gateway**: http://localhost:8000
- **Kong Admin**: http://localhost:8001
- **Konga UI**: http://localhost:1337
- **Kibana**: http://localhost:5601
- + Tất cả services như dev environment

## 🔧 Useful Commands

### View Logs
```powershell
# Development
docker-compose -f docker-compose.dev.yaml logs -f auth-service

# Production
docker-compose logs -f auth-service
```

### Rebuild Specific Service
```powershell
# Development
docker-compose -f docker-compose.dev.yaml up --build -d auth-service

# Production
docker-compose up --build -d auth-service
```

### Enter Container Shell
```powershell
docker exec -it golang-project-auth-service-1 sh
```

## 🔐 **Environment Variables & Secrets Management**

### All secrets are managed through individual .env files:
- **api-gateway/.env** - API Gateway configurations
- **auth-service/.env** - JWT secrets, Google OAuth, MongoDB
- **product-service/.env** - AWS S3 credentials, MongoDB, Redis
- **cart-service/.env** - MongoDB, Kafka, JWT
- **order-service/.env** - MongoDB, Kafka, SMTP settings
- **search-service/.env** - Elasticsearch, Redis settings

### Validate all .env files:
```powershell
.\validate-env.ps1
```

### Security Benefits:
- ✅ **No secrets in docker-compose files**
- ✅ **Each service manages its own secrets**
- ✅ **Easy to rotate individual secrets**
- ✅ **Different secrets per environment**
- ✅ **Git-ignored by default**

## 📁 File Structure

```
├── docker-compose.yaml          # Production environment
├── docker-compose.dev.yaml     # Development environment
├── start-dev.ps1               # Start development
├── start-prod.ps1              # Start production
├── stop-all.ps1                # Stop all services
│
├── auth-service/
│   ├── Dockerfile              # Production build
│   ├── Dockerfile.dev          # Development with Air
│   └── .air.toml               # Air configuration
│
├── product-service/
│   ├── Dockerfile
│   ├── Dockerfile.dev
│   └── .air.toml
│
└── ... (tương tự cho các service khác)
```

## 💡 Tips

1. **Sử dụng Development Environment** khi coding để tận hưởng hot reload
2. **Test trên Production Environment** trước khi deploy để đảm bảo tất cả hoạt động đúng
3. **Monitor logs** bằng Docker Desktop hoặc command line
4. **Clean up** định kỳ bằng `.\stop-all.ps1` để giải phóng resources
