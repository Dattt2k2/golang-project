# gRPC Order Service - Performance Improvements

## Vấn đề đã sửa
Order-service gRPC đang bị quá tải khi nhận nhiều request đồng thời, dẫn đến không nhận được request mới.

## Các thay đổi chính

### 1. **Sửa lỗi Registration Service** ✅
- **Trước**: Đăng ký `UnimplementedOrderServiceServer` - implementation rỗng không xử lý request
- **Sau**: Đăng ký `OrderServiceServer` thực tế với implementation đầy đủ
- **File**: `main.go` line 87-90

### 2. **Tăng cường Connection Pool Database** 🔧
- `MaxIdleConns`: 25 connections (giữ sẵn để xử lý nhanh)
- `MaxOpenConns`: 100 connections (tối đa đồng thời)
- `ConnMaxLifetime`: 1 giờ
- `ConnMaxIdleTime`: 10 phút
- **File**: `database/postgres.go`

### 3. **Cấu hình gRPC Server cho High Concurrency** ⚡
```go
- MaxConcurrentStreams: 1000      // Xử lý tối đa 1000 streams đồng thời
- MaxRecvMsgSize: 10MB            // Kích thước message tối đa
- MaxSendMsgSize: 10MB
- NumStreamWorkers: 100           // Tăng số workers
- KeepAlive parameters            // Tối ưu connection reuse
- EnforcementPolicy               // Chống resource exhaustion
```
**File**: `main.go` line 51-75

### 4. **Rate Limiting & Request Control** 🚦
- Semaphore giới hạn 500 concurrent requests
- Tự động từ chối request khi quá tải (ResourceExhausted error)
- Timeout mặc định 30 giây cho mỗi request
- **File**: `service/grpc_interceptor.go`

### 5. **Request Timeout & Context Handling** ⏱️
- Timeout 5 giây cho mỗi gRPC call
- Xử lý graceful timeout với channel
- Proper context cancellation
- **File**: `service/grpc_service.go`

### 6. **Health Check Service** 💚
- Endpoint health check cho monitoring
- Kiểm tra định kỳ trạng thái dependencies (Cart, Product service)
- Tự động update serving status
- **File**: `service/health_check.go`

### 7. **Request Logging & Monitoring** 📊
- Log mọi gRPC request với duration
- Track success/failure rate
- Cảnh báo khi rate limiting kick in
- **File**: `service/grpc_interceptor.go`

## Cách kiểm tra Health Check

```bash
# Sử dụng grpcurl để check health
grpcurl -plaintext localhost:8100 grpc.health.v1.Health/Check

# Với service name cụ thể
grpcurl -plaintext -d '{"service":"order.OrderService"}' localhost:8100 grpc.health.v1.Health/Check
```

## Metrics quan trọng cần monitor

1. **Connection Pool**
   - Số connections đang sử dụng
   - Số connections idle
   - Wait time cho connection

2. **gRPC Requests**
   - Request rate (req/s)
   - Response time (P50, P95, P99)
   - Error rate
   - Rate limiting rejections

3. **Resource Usage**
   - CPU usage
   - Memory usage
   - Goroutine count

## Load Testing

Để test hiệu năng, có thể dùng:

```bash
# Install ghz (gRPC load testing tool)
go install github.com/bojand/ghz/cmd/ghz@latest

# Test HasPurchased endpoint
ghz --insecure \
  --proto module/gRPC-Order/order_service.proto \
  --call order.OrderService/HasPurchased \
  -d '{"user_id":"test-user","product_id":"test-product"}' \
  -n 10000 \
  -c 100 \
  localhost:8100
```

## Khuyến nghị thêm

1. **Circuit Breaker**: Thêm circuit breaker pattern cho calls đến Cart/Product service
2. **Distributed Tracing**: Tích hợp OpenTelemetry để trace requests
3. **Metrics Export**: Export metrics sang Prometheus
4. **Horizontal Scaling**: Deploy nhiều instance của order-service với load balancer

## Environment Variables liên quan

```env
GRPC_PORT=8100              # gRPC server port
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password
POSTGRES_DB=order_db
```

## Troubleshooting

### Vấn đề: Vẫn bị quá tải
**Giải pháp**:
1. Tăng `MaxConcurrentStreams` trong main.go
2. Tăng semaphore limit trong `grpc_interceptor.go`
3. Scale horizontal với nhiều instances

### Vấn đề: Database connection pool exhausted
**Giải pháp**:
1. Tăng `MaxOpenConns` trong `database/postgres.go`
2. Kiểm tra xem có connection leak không (connections không được close)
3. Optimize queries để giảm thời gian giữ connection

### Vấn đề: Timeout errors
**Giải pháp**:
1. Tăng timeout trong `grpc_service.go` (hiện tại 5s)
2. Tối ưu database queries
3. Thêm caching layer

## Kết quả mong đợi

- ✅ Xử lý được hàng nghìn concurrent requests
- ✅ Response time ổn định dưới 100ms (P95)
- ✅ Không còn bị từ chối requests do quá tải
- ✅ Graceful degradation khi load cao
- ✅ Better observability với health checks và logging
