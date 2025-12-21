#!/bin/bash

# Script to run admin user migration
# This creates the admin user in both auth-service and user-service databases

set -e

echo "🚀 Running admin user migration..."

# Change to auth-service directory
cd "$(dirname "$0")/.." || exit

# Check if .env file exists (try multiple locations)
ENV_FOUND=false
if [ -f ".env" ]; then
    ENV_FOUND=true
    echo "✅ Found .env file in auth-service directory"
elif [ -f "../.env" ]; then
    ENV_FOUND=true
    echo "✅ Found .env file in parent directory"
elif [ -f "../../.env" ]; then
    ENV_FOUND=true
    echo "✅ Found .env file in root project directory"
fi

if [ "$ENV_FOUND" = false ]; then
    echo "⚠️  Warning: .env file not found in common locations"
    echo "   Tried: .env, ../.env, ../../.env"
    echo "   Continuing anyway - using environment variables if set"
fi

# Run the migration
go run cmd/migrate-admin/main.go

echo "✅ Migration completed!"

