#!/bin/bash
set -e

# Update system
apt-get update
apt-get upgrade -y

# Install required packages
apt-get install -y \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg \
    lsb-release \
    git \
    jq

# Install Docker
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
apt-get update
apt-get install -y docker-ce docker-ce-cli containerd.io

# Start and enable Docker
systemctl start docker
systemctl enable docker
groupadd -f docker
usermod -aG docker ubuntu

# Setup Docker logging
cat > /etc/docker/daemon.json << 'EOF'
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
EOF

systemctl restart docker

# Create Traefik configuration directory
mkdir -p /etc/traefik
chown ubuntu:ubuntu /etc/traefik

# Create Traefik static configuration
cat > /etc/traefik/traefik.yml << 'EOF'
global:
  checkNewVersion: true
  sendAnonymousUsage: false

api:
  dashboard: true
  insecure: true

entryPoints:
  web:
    address: ":80"
  websecure:
    address: ":443"
  traefik:
    address: ":8080"

ping:
  entryPoint: "web"

providers:
  file:
    filename: /etc/traefik/dynamic.yml
    watch: true

log:
  level: INFO
  format: json

accessLog:
  format: json
EOF

# Create Traefik dynamic configuration with service routing
cat > /etc/traefik/dynamic.yml << EOF
http:
  routers:
    auth-router:
      rule: "PathPrefix(\`/api/auth\`) || PathPrefix(\`/auth\`)"
      service: auth-service
      entryPoints:
        - web
    
    user-router:
      rule: "PathPrefix(\`/api/user\`) || PathPrefix(\`/user\`)"
      service: user-service
      entryPoints:
        - web
    
    product-router:
      rule: "PathPrefix(\`/api/product\`) || PathPrefix(\`/product\`)"
      service: product-service
      entryPoints:
        - web
    
    cart-router:
      rule: "PathPrefix(\`/api/cart\`) || PathPrefix(\`/cart\`)"
      service: cart-service
      entryPoints:
        - web
    
    order-router:
      rule: "PathPrefix(\`/api/order\`) || PathPrefix(\`/order\`)"
      service: order-service
      entryPoints:
        - web
    
    payment-router:
      rule: "PathPrefix(\`/api/payment\`) || PathPrefix(\`/payment\`)"
      service: payment-service
      entryPoints:
        - web

  services:
    auth-service:
      loadBalancer:
        servers:
          - url: "http://${AUTH_SERVICE_IP}:8081"
    
    user-service:
      loadBalancer:
        servers:
          - url: "http://${USER_SERVICE_IP}:8082"
    
    product-service:
      loadBalancer:
        servers:
          - url: "http://${PRODUCT_SERVICE_IP}:8083"
    
    cart-service:
      loadBalancer:
        servers:
          - url: "http://${CART_SERVICE_IP}:8085"
    
    order-service:
      loadBalancer:
        servers:
          - url: "http://${ORDER_SERVICE_IP}:8084"
    
    payment-service:
      loadBalancer:
        servers:
          - url: "http://${PAYMENT_SERVICE_IP}:8086"
EOF

# Run Traefik as Docker container
docker run -d \
  --name traefik \
  --restart unless-stopped \
  -p 80:80 \
  -p 443:443 \
  -p 8080:8080 \
  -v /etc/traefik:/etc/traefik:ro \
  traefik:v2.10

# Create systemd service
cat > /etc/systemd/system/traefik.service << 'EOF'
[Unit]
Description=Traefik API Gateway
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
ExecStart=/usr/bin/docker start traefik
ExecStop=/usr/bin/docker stop traefik
User=root

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable traefik.service

# Setup firewall
ufw --force enable
ufw default deny incoming
ufw default allow outgoing
ufw allow 22/tcp
ufw allow 80/tcp
ufw allow 443/tcp
ufw allow 8080/tcp

echo "Traefik setup completed!"
