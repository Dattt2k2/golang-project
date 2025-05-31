#!/bin/bash

echo "=== Kong Setup Script ==="
echo "Setting up Kong services and routes..."

# Kiểm tra Kong status
echo "Checking Kong status..."
until curl -f http://kong:8001/status 2>/dev/null; do
    echo "Kong is not ready yet... waiting 5 seconds"
    sleep 5
done
echo "Kong is ready!"

# Add Services
echo "Adding services..."

echo "Adding Auth Service..."
curl -s -X POST http://kong:8001/services/ \
  --data "name=auth-service" \
  --data "url=http://auth-service:8081"

echo "Adding Product Service..."
curl -s -X POST http://kong:8001/services/ \
  --data "name=product-service" \
  --data "url=http://product-service:8082"

echo "Adding Search Service..."
curl -s -X POST http://kong:8001/services/ \
  --data "name=search-service" \
  --data "url=http://search-service:8086"



echo "Adding Cart Service..."
curl -s -X POST http://kong:8001/services/ \
  --data "name=cart-service" \
  --data "url=http://cart-service:8083"

echo "Adding Order Service..."
curl -s -X POST http://kong:8001/services/ \
  --data "name=order-service" \
  --data "url=http://order-service:8084"

# Add Routes
echo "Adding routes..."

curl -s -X POST http://kong:8001/services/auth-service/routes \
  --data "name=auth-routes" \
  --data "paths[]=/api/auth" \
  --data "strip_path=false"

# Product public routes (no auth needed)
curl -s -X POST http://kong:8001/services/product-service/routes \
  --data "name=product-public-routes" \
  --data "paths[]=/api/products/get" \
  --data "paths[]=/api/products/search" \
  --data "paths[]=/api/best-selling" \
  --data "strip_path=false"

# Product protected routes (auth needed) 
curl -s -X POST http://kong:8001/services/product-service/routes \
  --data "name=product-protected-routes" \
  --data "paths[]=/api/products/add" \
  --data "paths[]=/api/products/edit" \
  --data "paths[]=/api/products/delete" \
  --data "strip_path=false"

curl -s -X POST http://kong:8001/services/search-service/routes \
  --data "name=search-routes" \
  --data "paths[]=/api/search" \
  --data "strip_path=false"



curl -s -X POST http://kong:8001/services/cart-service/routes \
  --data "name=cart-routes" \
  --data "paths[]=/api/cart" \
  --data "strip_path=false"

curl -s -X POST http://kong:8001/services/order-service/routes \
  --data "name=order-routes" \
  --data "paths[]=/api/orders" \
  --data "strip_path=false"

# Add Global Plugins
echo "Adding plugins..."

# CORS Plugin
echo "Adding CORS plugin..."
curl -s -X POST http://kong:8001/plugins/ \
  --data "name=cors" \
  --data "config.origins=*" \
  --data "config.methods=GET,POST,PUT,DELETE,PATCH,OPTIONS" \
  --data "config.headers=Accept,Accept-Version,Content-Length,Content-MD5,Content-Type,Date,Authorization,X-User-ID" \
  --data "config.exposed_headers=X-User-ID" \
  --data "config.credentials=true" \
  --data "config.max_age=3600"

# Rate Limiting Plugin
echo "Adding Rate Limiting plugin..."
curl -s -X POST http://kong:8001/plugins/ \
  --data "name=rate-limiting" \
  --data "config.minute=100" \
  --data "config.hour=1000" \
  --data "config.policy=local"

# Request Size Limiting
echo "Adding Request Size Limiting plugin..."
curl -s -X POST http://kong:8001/plugins/ \
  --data "name=request-size-limiting" \
  --data "config.allowed_payload_size=128"

# JWT Plugin for protected routes  
echo "Setting up JWT authentication for protected routes..."

# Create JWT Consumer for auth-service
echo "Creating JWT consumer..."
curl -s -X POST http://kong:8001/consumers/ \
  --data "username=auth-service-consumer"

# Add JWT credentials to consumer with the same secret key used by auth-service
echo "Adding JWT credentials..."
curl -s -X POST http://kong:8001/consumers/auth-service-consumer/jwt \
  --data "algorithm=HS256" \
  --data "key=golang-project" \
  --data "secret=12dat34"

# Add JWT plugin to protected product routes
echo "Adding JWT plugin for product protected routes..."
curl -s -X POST http://kong:8001/routes/product-protected-routes/plugins/ \
  --data "name=jwt" \
  --data "config.secret_is_base64=false"

# Add Request Transformer plugin to forward user_id header for product routes
echo "Adding Request Transformer for product protected routes..."
curl -s -X POST http://kong:8001/routes/product-protected-routes/plugins/ \
  --data "name=request-transformer" \
  --data "config.add.headers[]=user_id:\$(jwt.claims.uid)" \
  --data "config.add.headers[]=user_type:\$(jwt.claims.user_type)"

# Add JWT plugin to cart service
echo "Adding JWT plugin for cart service..."
curl -s -X POST http://kong:8001/services/cart-service/plugins/ \
  --data "name=jwt" \
  --data "config.secret_is_base64=false"

# Add Request Transformer plugin to forward user_id header for cart service
echo "Adding Request Transformer for cart service..."
curl -s -X POST http://kong:8001/services/cart-service/plugins/ \
  --data "name=request-transformer" \
  --data "config.add.headers[]=user_id:\$(jwt.claims.uid)" \
  --data "config.add.headers[]=user_type:\$(jwt.claims.user_type)"

# Add JWT plugin to order service  
echo "Adding JWT plugin for order service..."
curl -s -X POST http://kong:8001/services/order-service/plugins/ \
  --data "name=jwt" \
  --data "config.secret_is_base64=false"

# Add Request Transformer plugin to forward user_id header for order service
echo "Adding Request Transformer for order service..."
curl -s -X POST http://kong:8001/services/order-service/plugins/ \
  --data "name=request-transformer" \
  --data "config.add.headers[]=user_id:\$(jwt.claims.uid)" \
  --data "config.add.headers[]=user_type:\$(jwt.claims.user_type)"

echo "=== Kong setup completed successfully! ==="
echo "Kong Proxy: http://localhost:8000"
echo "Kong Admin: http://localhost:8001"
echo "Konga UI: http://localhost:1337"
echo ""
echo "API Endpoints:"
echo "- Auth: http://localhost:8000/api/auth/*"
echo "- Products: http://localhost:8000/api/products/*"
echo "- Search: http://localhost:8000/api/search/*"
echo "- Cart: http://localhost:8000/api/cart/*"
echo "- Orders: http://localhost:8000/api/orders/*"