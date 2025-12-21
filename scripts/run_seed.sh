#!/bin/bash

# Script to run seed data
# Make sure you're in the project root directory

echo "🌱 Starting seed data script..."

# Check if we're in the right directory
if [ ! -f "scripts/seed_data.go" ]; then
    echo "❌ Error: seed_data.go not found. Please run this script from the project root directory."
    exit 1
fi

# Navigate to scripts directory
cd scripts

# Install dependencies
echo "📦 Installing dependencies..."
go mod tidy

# Check if .env file exists in parent directory
if [ -f "../.env" ]; then
    echo "✅ Found .env file"
else
    echo "⚠️  Warning: .env file not found. Using environment variables or defaults."
fi

# Run the seed script
echo "🚀 Running seed script..."
go run seed_data.go

if [ $? -eq 0 ]; then
    echo "✅ Seed data completed successfully!"
else
    echo "❌ Seed data failed. Please check the error messages above."
    exit 1
fi

