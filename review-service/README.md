# Review Service

Service quản lý đánh giá sản phẩm trong hệ thống thương mại điện tử.

## Công nghệ
- Go 1.23
- Gin framework
- Redis (lưu trữ)
- Air (hot reload)

## Cấu trúc thư mục
```
review-service/
├── cmd/server/          # Entry point
├── internal/
│   ├── models/         # Data models
│   ├── repository/     # Data access layer
│   ├── handlers/       # HTTP handlers
│   └── routes/         # Route definitions
├── Dockerfile.dev      # Development dockerfile
├── .air.toml          # Air config
└── .env               # Environment variables
```

## Endpoints

### Health Check
- `GET /health` - Kiểm tra trạng thái service

### Reviews
- `POST /v1/products/:product_id/reviews` - Tạo review mới
- `GET /v1/products/:product_id/reviews` - Lấy danh sách reviews của sản phẩm
- `GET /v1/reviews/:review_id` - Lấy chi tiết review

## Cấu hình môi trường

```properties
PORT=8089
REDIS_ADDR=redis:6379
REDIS_DB=0
REDIS_PASSWORD=
SERVICE_NAME=review-service
```

## Chạy local

```bash
# Với docker-compose
docker compose up -d review-service

# Logs
docker compose logs -f review-service
```

## API Examples

### Tạo review
```bash
curl -X POST http://localhost:8092/v1/products/product123/reviews \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "rating": 5,
    "title": "Sản phẩm tuyệt vời",
    "body": "Chất lượng rất tốt, giao hàng nhanh"
  }'
```

### Lấy reviews
```bash
curl http://localhost:8092/v1/products/product123/reviews?limit=10
```
