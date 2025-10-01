# User Service

This project is a user service built in Go, designed to handle user-related operations such as creating, retrieving, updating, and deleting users. It utilizes a clean architecture approach, separating concerns into different packages.

## Project Structure

```
user-service
├── cmd
│   └── server
│       └── main.go          # Entry point of the application
├── internal
│   ├── handlers
│   │   └── users.go         # HTTP handlers for user-related endpoints
│   ├── services
│   │   └── users_service.go  # Business logic for user operations
│   ├── repository
│   │   └── users_repo.go     # Database interaction methods
│   ├── models
│   │   └── users_model.go     # User entity definition
│   └── config
│       └── config.go         # Application configuration handling
├── migrations
│   └── 0001_create_users_table.up.sql  # SQL migration script for users table
├── scripts
│   └── migrate.sh            # Shell script to run database migrations
├── configs
│   └── config.yaml           # Configuration settings in YAML format
├── Dockerfile                 # Docker image build file
├── Makefile                   # Task definitions for building and running the application
├── go.mod                     # Module definition and dependencies
├── go.sum                     # Checksums for module dependencies
├── .env.example               # Example environment variables
└── README.md                  # Project documentation
```

## Setup Instructions

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd user-service
   ```

2. **Install dependencies:**
   ```bash
   go mod tidy
   ```

3. **Configure environment variables:**
   
   Copy `.env.example` to `.env` and update the values:
   ```bash
   cp .env.example .env
   ```
   
   Required environment variables:
   ```bash
   DB_HOST=localhost          # Database host
   DB_PORT=5432              # Database port
   DB_USER=postgres          # Database user
   DB_PASSWORD=your_password # Database password
   DB_NAME=user_service_db   # Database name
   SERVER_HOST=0.0.0.0       # Server host
   SERVER_PORT=8080          # Server port
   ```

4. **Setup PostgreSQL database:**
   ```bash
   # Create database
   createdb user_service_db
   
   # Or run migrations manually
   psql -U postgres -d user_service_db -f migrations/0001_create_users_table.up.sql
   ```

5. **Start the application:**
   ```bash
   # Load environment variables and run
   export $(cat .env | xargs) && go run cmd/server/main.go
   
   # Or use make (if Makefile is configured)
   make run
   ```

## API Endpoints

The service exposes the following REST API endpoints:

- `POST /users` - Create a new user
- `GET /users/:id` - Get user by ID
- `PUT /users/:id` - Update user by ID
- `DELETE /users/:id` - Delete user by ID

## Usage

- The service exposes various HTTP endpoints for user operations. Refer to the `internal/handlers/users.go` file for the list of available endpoints and their usage.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any enhancements or bug fixes.