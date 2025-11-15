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
    unzip \
    jq

# Install Docker
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
apt-get update
apt-get install -y docker-ce docker-ce-cli containerd.io

# Start and enable Docker
systemctl start docker
systemctl enable docker

# Install Docker Compose
curl -L "https://github.com/docker/compose/releases/download/v${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# Create docker group and add ubuntu user
groupadd -f docker
usermod -aG docker ubuntu

# Create application directory
mkdir -p /opt/microservices
chown ubuntu:ubuntu /opt/microservices

# Create environment file placeholder
cat > /opt/microservices/.env << 'EOF'
# Database Configuration
AUTH_DB_HOST=${AUTH_DB_HOST}
AUTH_DB_PORT=5432
AUTH_DB_NAME=${AUTH_DB_NAME}
AUTH_DB_USER=${DB_USERNAME}
AUTH_DB_PASSWORD=${DB_PASSWORD}

USER_DB_HOST=${USER_DB_HOST}
USER_DB_PORT=5432
USER_DB_NAME=${USER_DB_NAME}
USER_DB_USER=${DB_USERNAME}
USER_DB_PASSWORD=${DB_PASSWORD}

PAYMENT_DB_HOST=${PAYMENT_DB_HOST}
PAYMENT_DB_PORT=5432
PAYMENT_DB_NAME=${PAYMENT_DB_NAME}
PAYMENT_DB_USER=${DB_USERNAME}
PAYMENT_DB_PASSWORD=${DB_PASSWORD}

ORDER_DB_HOST=${ORDER_DB_HOST}
ORDER_DB_PORT=5432
ORDER_DB_NAME=${ORDER_DB_NAME}
ORDER_DB_USER=${DB_USERNAME}
ORDER_DB_PASSWORD=${DB_PASSWORD}

# Redis Configuration
REDIS_HOST=redis
REDIS_PORT=6379

# Kafka Configuration
KAFKA_BROKER=kafka:9092

# Elasticsearch Configuration
ELASTICSEARCH_URL=http://elasticsearch:9200

# Application Configuration
ENVIRONMENT=production
LOG_LEVEL=info

# Add your Stripe keys and other secrets here
STRIPE_SECRET_KEY=your_stripe_secret_key
STRIPE_WEBHOOK_SECRET=your_webhook_secret
JWT_SECRET=your_jwt_secret

EOF

chown ubuntu:ubuntu /opt/microservices/.env

# Create systemd service for docker-compose
cat > /etc/systemd/system/microservices.service << 'EOF'
[Unit]
Description=Microservices Application
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/opt/microservices
ExecStart=/usr/local/bin/docker-compose up -d
ExecStop=/usr/local/bin/docker-compose down
User=ubuntu

[Install]
WantedBy=multi-user.target
EOF

# Enable the service
systemctl daemon-reload
systemctl enable microservices.service

# Install monitoring tools (optional)
apt-get install -y htop iotop nethogs

# Setup log rotation for Docker
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

# Setup firewall (UFW)
ufw --force enable
ufw default deny incoming
ufw default allow outgoing
ufw allow 22/tcp
ufw allow 80/tcp
ufw allow 443/tcp
ufw allow 8080/tcp

# Create deployment script
cat > /opt/microservices/deploy.sh << 'DEPLOY_SCRIPT'
#!/bin/bash
set -e

echo "Pulling latest images..."
docker-compose pull

echo "Stopping services..."
docker-compose down

echo "Starting services..."
docker-compose up -d

echo "Waiting for services to be healthy..."
sleep 10

echo "Service status:"
docker-compose ps

echo "Deployment complete!"
DEPLOY_SCRIPT

chmod +x /opt/microservices/deploy.sh
chown ubuntu:ubuntu /opt/microservices/deploy.sh

echo "User data script completed successfully!"
