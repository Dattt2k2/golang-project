#!/bin/bash

# Script to run admin user migration
# This creates the admin user in both auth-service and user-service databases

set -e

echo "🚀 Running admin user migration..."

# Change to auth-service directory
cd "$(dirname "$0")/.." || exit

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo "⚠️  Warning: .env file not found in auth-service directory"
    echo "   Trying to load from parent directory..."
    if [ ! -f "../.env" ]; then
        echo "❌ Error: .env file not found. Please create one or set environment variables."
        exit 1
    fi
fi

# Run the migration
go run cmd/migrate-admin/main.go

echo "✅ Migration completed!"
