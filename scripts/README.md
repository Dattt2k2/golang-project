# Seed Data Script

Script này được sử dụng để seed dữ liệu test cho hệ thống:
- Tạo 20 users
- Lấy products từ product-service
- Tạo orders cho tất cả 12 tháng trong năm 2025 (mỗi tháng tối thiểu 10 orders)

## Tạo Tài Khoản Admin

**LƯU Ý:** Script seed data không tạo tài khoản admin. Để tạo tài khoản admin, bạn cần chạy migration riêng.

### Cách 1: Sử dụng shell script (Khuyến nghị)

**Lưu ý:** Nếu gặp lỗi "permission denied", bạn cần cấp quyền thực thi cho script trước:

```bash
cd auth-service
chmod +x scripts/migrate-admin.sh
./scripts/migrate-admin.sh
```

Hoặc chạy trực tiếp với bash:
```bash
cd auth-service
bash scripts/migrate-admin.sh
```

### Cách 2: Chạy trực tiếp với Go

```bash
cd auth-service
go run cmd/migrate-admin/main.go
```

### Cách 3: Build và chạy

```bash
cd auth-service
go build -o bin/migrate-admin cmd/migrate-admin/main.go
./bin/migrate-admin
```

### Thông tin Tài Khoản Admin

Sau khi chạy migration, tài khoản admin sẽ được tạo với thông tin:
- **Email**: `admin@ecomo.com`
- **Password**: `Aminumina1!`
- **Name**: `Admin`
- **User Type**: `ADMIN`
- **Is Verified**: `true`

### Yêu cầu cho Admin Migration

Migration cần các biến môi trường sau (có thể đặt trong file `.env` ở root project):

**Auth Database:**
- `AUTH_POSTGRES_HOST` (mặc định: `localhost`)
- `AUTH_POSTGRES_PORT` (mặc định: `5432`)
- `AUTH_POSTGRES_USER` (mặc định: `postgres`)
- `AUTH_POSTGRES_PASSWORD` (**BẮT BUỘC**)
- `AUTH_POSTGRES_DB` (mặc định: `auth_db`)

**User Database (tùy chọn - để tạo admin trong user-service):**
- `USER_POSTGRES_HOST` (mặc định: `localhost`)
- `USER_POSTGRES_PORT` (mặc định: `5433`)
- `USER_POSTGRES_USER` (mặc định: `postgres`)
- `USER_POSTGRES_PASSWORD` (**BẮT BUỘC**)
- `USER_POSTGRES_DB` (mặc định: `user_db`)

**Lưu ý:**
- Migration là idempotent - có thể chạy nhiều lần an toàn
- Nếu admin user đã tồn tại, migration sẽ cập nhật thông tin (password, name, etc.)
- Nếu không kết nối được user-service database, admin vẫn sẽ được tạo trong auth-service

## Yêu cầu

1. Go 1.24.4 hoặc cao hơn
2. PostgreSQL databases cho auth-service và order-service đang chạy
3. Product-service đang chạy và có sẵn products

## Cấu hình

**QUAN TRỌNG:** Script cần các biến môi trường để kết nối database. Bạn có thể:
1. Tạo file `.env` ở root project với các biến môi trường (xem `scripts/.env.example`)
2. Hoặc export các biến môi trường trước khi chạy script

**Cách nhanh nhất:**
```bash
# Copy file example
cp scripts/.env.example .env

# Chỉnh sửa file .env với thông tin database của bạn
nano .env  # hoặc vim .env

# Chạy script
cd scripts && go run seed_data.go
```

Script sẽ đọc các biến môi trường sau (có thể đặt trong file `.env` ở root project):

### Auth Database
- `AUTH_POSTGRES_HOST` (mặc định: `localhost`)
- `AUTH_POSTGRES_PORT` (mặc định: `5432`)
- `AUTH_POSTGRES_USER` (mặc định: `POSTGRES_USER` hoặc `postgres`)
- `AUTH_POSTGRES_PASSWORD` (**BẮT BUỘC** - hoặc dùng `POSTGRES_PASSWORD`)
- `AUTH_POSTGRES_DB` (mặc định: `POSTGRES_DB` hoặc `authdb`)
- `AUTH_POSTGRES_SSLMODE` (mặc định: `disable`)

### User Database (for addresses)
- `USER_POSTGRES_HOST` (mặc định: `POSTGRES_HOST` hoặc `localhost`)
- `USER_POSTGRES_PORT` (mặc định: `POSTGRES_PORT` hoặc `5432`)
- `USER_POSTGRES_USER` (mặc định: `POSTGRES_USER` hoặc `postgres`)
- `USER_POSTGRES_PASSWORD` (**BẮT BUỘC** - hoặc dùng `POSTGRES_PASSWORD`)
- `USER_POSTGRES_DB` (mặc định: `DB_NAME` hoặc `user_service_db`)
- `USER_POSTGRES_SSLMODE` (mặc định: `disable`)

### Order Database
- `ORDER_POSTGRES_HOST` (mặc định: `POSTGRES_HOST` hoặc `localhost`)
- `ORDER_POSTGRES_PORT` (mặc định: `POSTGRES_PORT` hoặc `5432`)
- `ORDER_POSTGRES_USER` (mặc định: `POSTGRES_USER` hoặc `postgres`)
- `ORDER_POSTGRES_PASSWORD` (**BẮT BUỘC** - hoặc dùng `POSTGRES_PASSWORD`)
- `ORDER_POSTGRES_DB` (mặc định: `POSTGRES_DB` hoặc `orderdb`)
- `ORDER_POSTGRES_SSLMODE` (mặc định: `disable`)

### Product Service
- `PRODUCT_SERVICE_URL` (mặc định: `http://localhost:8082`)

## Cách chạy Seed Data

### Cách 1: Chạy trực tiếp với Go

```bash
cd scripts
go mod tidy
go run seed_data.go
```

### Cách 2: Build và chạy

```bash
cd scripts
go mod tidy
go build -o seed_data seed_data.go
./seed_data
```

### Cách 3: Sử dụng script shell

```bash
chmod +x scripts/run_seed.sh
./scripts/run_seed.sh
```

## Thông tin Users được tạo

- Email: `user1@example.com` đến `user20@example.com`
- Password: `password123` (cho tất cả users - có thể thay đổi qua `SEED_USER_PASSWORD`)
- Phone: `0900000001` đến `0900000020` (format Việt Nam)
- Tất cả users đều được verify (`IsVerify = true`)
- User type: `USER`
- Mỗi user có một địa chỉ mặc định được tạo trong bảng `user_addresses` của user-service database

## Thông tin Orders

- Mỗi tháng sẽ có 10-15 orders (ngẫu nhiên)
- Mỗi order có 1-5 items (ngẫu nhiên)
- Orders được phân bố ngẫu nhiên trong tháng
- Payment methods: COD hoặc STRIPE (ngẫu nhiên)
- Status phụ thuộc vào payment method
- Shipping address được lấy từ địa chỉ mặc định của user trong bảng `user_addresses`

## Lưu ý

- Script sẽ bỏ qua các products không có variants hoặc status không phải "onsale"
- Script sẽ bỏ qua các variants không có stock
- Nếu user hoặc order tạo thất bại, script sẽ log warning và tiếp tục
- Đảm bảo product-service đang chạy và có ít nhất một số products trước khi chạy script

## Truy cập Product Service

Script sẽ cố gắng truy cập product-service theo thứ tự:
1. `PRODUCT_SERVICE_URL` environment variable (nếu có)
2. `http://product-service:8082` (nếu chạy trong Docker network)
3. `http://localhost:8082` (nếu chạy local)

**Lưu ý về Authentication:**
- Nếu product-service yêu cầu authentication, bạn có thể:
  1. Chạy script trong cùng Docker network với product-service (khuyến nghị)
  2. Hoặc sửa code để thêm authentication token vào request headers
  3. Hoặc tạm thời disable authentication cho endpoint `/products/get/all` trong product-service (chỉ cho môi trường dev)

## Troubleshooting

### Lỗi "password authentication failed"
**Nguyên nhân:** Thiếu hoặc sai password cho database.

**Giải pháp:**
1. Tạo file `.env` ở root project với nội dung:
```bash
# Database passwords (BẮT BUỘC)
POSTGRES_PASSWORD=your_password_here

# Hoặc chỉ định riêng cho từng service
AUTH_POSTGRES_PASSWORD=your_auth_db_password
USER_POSTGRES_PASSWORD=your_user_db_password
ORDER_POSTGRES_PASSWORD=your_order_db_password

# Database hosts (nếu khác localhost)
AUTH_POSTGRES_HOST=localhost
USER_POSTGRES_HOST=localhost
ORDER_POSTGRES_HOST=localhost

# Database names
AUTH_POSTGRES_DB=authdb
USER_POSTGRES_DB=user_service_db
ORDER_POSTGRES_DB=orderdb
```

2. Hoặc export trước khi chạy:
```bash
export POSTGRES_PASSWORD=your_password
cd scripts && go run seed_data.go
```

### Lỗi "authentication required" (product-service)
- Đảm bảo script đang chạy trong cùng Docker network với product-service
- Hoặc thêm authentication token vào script

### Lỗi "no products found"
- Kiểm tra product-service có đang chạy không
- Kiểm tra có products nào trong database không
- Kiểm tra products có status "onsale" và có variants không

