#!/bin/bash

# Script to verify admin user exists in both databases

echo "🔍 Verifying admin user in databases..."
echo ""

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
    echo "✅ Loaded .env file"
elif [ -f ../.env ]; then
    export $(cat ../.env | grep -v '^#' | xargs)
    echo "✅ Loaded .env file from parent directory"
elif [ -f ../../.env ]; then
    export $(cat ../../.env | grep -v '^#' | xargs)
    echo "✅ Loaded .env file from grandparent directory"
else
    echo "⚠️  No .env file found"
fi

# Get database connection info
AUTH_HOST=${AUTH_POSTGRES_HOST:-${POSTGRES_HOST:-localhost}}
AUTH_PORT=${AUTH_POSTGRES_PORT:-${POSTGRES_PORT:-5432}}
AUTH_USER=${AUTH_POSTGRES_USER:-${POSTGRES_USER:-postgres}}
AUTH_PASSWORD=${AUTH_POSTGRES_PASSWORD:-${POSTGRES_PASSWORD}}
AUTH_DB=${AUTH_POSTGRES_DB:-${POSTGRES_DB:-auth_db}}

USER_HOST=${USER_POSTGRES_HOST:-${POSTGRES_HOST:-localhost}}
USER_PORT=${USER_POSTGRES_PORT:-${POSTGRES_PORT:-5433}}
USER_USER=${USER_POSTGRES_USER:-${POSTGRES_USER:-postgres}}
USER_PASSWORD=${USER_POSTGRES_PASSWORD:-${POSTGRES_PASSWORD}}
USER_DB=${USER_POSTGRES_DB:-${POSTGRES_DB:-user_db}}

ADMIN_EMAIL="admin@ecomo.com"

# Try localhost if Docker hostname fails
if [ "${AUTH_HOST}" != "localhost" ] && [ "${AUTH_HOST}" != "127.0.0.1" ]; then
    AUTH_HOST_TRY="${AUTH_HOST}"
else
    AUTH_HOST_TRY="localhost"
fi

if [ "${USER_HOST}" != "localhost" ] && [ "${USER_HOST}" != "127.0.0.1" ]; then
    USER_HOST_TRY="${USER_HOST}"
else
    USER_HOST_TRY="localhost"
fi

echo ""
echo "📊 Checking auth-service database (${AUTH_HOST_TRY}:${AUTH_PORT}/${AUTH_DB})..."
PGPASSWORD="${AUTH_PASSWORD}" psql -h "${AUTH_HOST_TRY}" -p "${AUTH_PORT}" -U "${AUTH_USER}" -d "${AUTH_DB}" -c "
SELECT id, email, first_name, last_name, user_type, is_verify, created_at, updated_at 
FROM users 
WHERE email = '${ADMIN_EMAIL}';
" 2>/dev/null || {
    if [ "${AUTH_HOST_TRY}" != "localhost" ]; then
        echo "⚠️  Failed to connect to ${AUTH_HOST_TRY}, trying localhost..."
        PGPASSWORD="${AUTH_PASSWORD}" psql -h "localhost" -p "${AUTH_PORT}" -U "${AUTH_USER}" -d "${AUTH_DB}" -c "
SELECT id, email, first_name, last_name, user_type, is_verify, created_at, updated_at 
FROM users 
WHERE email = '${ADMIN_EMAIL}';
" 2>/dev/null || echo "❌ Failed to connect to auth-service database"
    else
        echo "❌ Failed to connect to auth-service database"
    fi
}

echo ""
echo "📊 Checking user-service database (${USER_HOST_TRY}:${USER_PORT}/${USER_DB})..."
PGPASSWORD="${USER_PASSWORD}" psql -h "${USER_HOST_TRY}" -p "${USER_PORT}" -U "${USER_USER}" -d "${USER_DB}" -c "
SELECT id, email, first_name, last_name, user_type, created_at, updated_at, deleted_at 
FROM users 
WHERE email = '${ADMIN_EMAIL}';
" 2>/dev/null || {
    if [ "${USER_HOST_TRY}" != "localhost" ]; then
        echo "⚠️  Failed to connect to ${USER_HOST_TRY}, trying localhost..."
        PGPASSWORD="${USER_PASSWORD}" psql -h "localhost" -p "${USER_PORT}" -U "${USER_USER}" -d "${USER_DB}" -c "
SELECT id, email, first_name, last_name, user_type, created_at, updated_at, deleted_at 
FROM users 
WHERE email = '${ADMIN_EMAIL}';
" 2>/dev/null || echo "❌ Failed to connect to user-service database"
    else
        echo "❌ Failed to connect to user-service database"
    fi
}

echo ""
echo "📊 Checking user-service database (including soft-deleted)..."
PGPASSWORD="${USER_PASSWORD}" psql -h "${USER_HOST_TRY}" -p "${USER_PORT}" -U "${USER_USER}" -d "${USER_DB}" -c "
SELECT id, email, first_name, last_name, user_type, created_at, updated_at, deleted_at 
FROM users 
WHERE email = '${ADMIN_EMAIL}' OR email LIKE '%admin%';
" 2>/dev/null || {
    if [ "${USER_HOST_TRY}" != "localhost" ]; then
        echo "⚠️  Failed to connect to ${USER_HOST_TRY}, trying localhost..."
        PGPASSWORD="${USER_PASSWORD}" psql -h "localhost" -p "${USER_PORT}" -U "${USER_USER}" -d "${USER_DB}" -c "
SELECT id, email, first_name, last_name, user_type, created_at, updated_at, deleted_at 
FROM users 
WHERE email = '${ADMIN_EMAIL}' OR email LIKE '%admin%';
" 2>/dev/null || echo "❌ Failed to connect to user-service database"
    else
        echo "❌ Failed to connect to user-service database"
    fi
}

echo ""
echo "✅ Verification complete"

