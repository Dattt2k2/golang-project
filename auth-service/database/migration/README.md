# Database Migrations

This directory contains database migration scripts for the auth-service.

## Admin User Migration

The `create_admin_user.go` migration creates an admin user in both auth-service and user-service databases.

### Admin User Details

- **Email**: `admin@ecomo.com`
- **Password**: `Aminumina1!`
- **Name**: `Admin`
- **User Type**: `ADMIN`
- **Is Verified**: `true`

### Running the Migration

#### Option 1: Using the shell script (Recommended)

**Lưu ý:** Nếu gặp lỗi "permission denied", bạn cần cấp quyền thực thi cho script:

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

#### Option 2: Using Go directly

```bash
cd auth-service
go run cmd/migrate-admin/main.go
```

#### Option 3: Build and run

```bash
cd auth-service
go build -o bin/migrate-admin cmd/migrate-admin/main.go
./bin/migrate-admin
```

### Troubleshooting

**Lỗi "lookup postgres-auth: no such host":**
- Nếu bạn chạy migration từ local (không trong Docker network), nhưng file `.env` có cấu hình Docker service names (như `postgres-auth`), migration sẽ tự động thử kết nối với `localhost` như fallback.
- Hoặc bạn có thể set environment variables trực tiếp:
  ```bash
  export AUTH_POSTGRES_HOST=localhost
  export AUTH_POSTGRES_PORT=5432
  export AUTH_POSTGRES_PASSWORD=your_password
  go run cmd/migrate-admin/main.go
  ```

### Environment Variables Required

The migration requires the following environment variables:

**For auth-service database:**
- `AUTH_POSTGRES_HOST` (or `POSTGRES_HOST`) - Default: `localhost`
- `AUTH_POSTGRES_PORT` (or `POSTGRES_PORT`) - Default: `5432`
- `AUTH_POSTGRES_USER` (or `POSTGRES_USER`) - Default: `postgres`
- `AUTH_POSTGRES_PASSWORD` (or `POSTGRES_PASSWORD`) - **Required**
- `AUTH_POSTGRES_DB` (or `POSTGRES_DB`) - Default: `auth_db`
- `AUTH_POSTGRES_SSLMODE` (or `POSTGRES_SSLMODE`) - Default: `disable`

**For user-service database (optional):**
- `USER_POSTGRES_HOST` (or `POSTGRES_HOST`) - Default: `localhost`
- `USER_POSTGRES_PORT` (or `POSTGRES_PORT`) - Default: `5433`
- `USER_POSTGRES_USER` (or `POSTGRES_USER`) - Default: `postgres`
- `USER_POSTGRES_PASSWORD` (or `POSTGRES_PASSWORD`) - **Required**
- `USER_POSTGRES_DB` (or `POSTGRES_DB`) - Default: `user_db`
- `USER_POSTGRES_SSLMODE` (or `POSTGRES_SSLMODE`) - Default: `disable`

### Behavior

- If the admin user already exists, the migration will update the user's information (password, name, etc.)
- The migration will create the admin user in both auth-service and user-service databases
- If user-service database connection fails, the migration will still create the admin user in auth-service

### Integration with Main Service

You can optionally integrate this migration into the main service startup by calling it in `main.go`:

```go
import "auth-service/database/migration"

func main() {
    // ... existing code ...
    
    // Run admin user migration
    if err := migration.CreateAdminUserMigration(); err != nil {
        logger.Logger.Warn("Failed to create admin user", logger.ErrField(err))
        // Don't fail startup if migration fails
    }
    
    // ... rest of the code ...
}
```

### Notes

- The migration is idempotent - it can be run multiple times safely
- If the admin user exists, it will update the password and other fields
- The password is hashed using bcrypt with cost factor 14

