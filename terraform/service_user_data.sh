#!/bin/bash
set -e

# Service configuration
SERVICE_NAME="${SERVICE_NAME}"
SERVICE_PORT="${SERVICE_PORT}"

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
    jq \
    htop

# Install Docker
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
apt-get update
apt-get install -y docker-ce docker-ce-cli containerd.io

# Start and enable Docker
systemctl start docker
systemctl enable docker

# Create docker group and add ubuntu user
groupadd -f docker
usermod -aG docker ubuntu

# Create application directory
mkdir -p /opt/$SERVICE_NAME
chown ubuntu:ubuntu /opt/$SERVICE_NAME

# Create environment file
cat > /opt/$SERVICE_NAME/.env << 'EOF'
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

# Infrastructure
REDIS_HOST=${REDIS_HOST}
REDIS_PORT=6379

KAFKA_BROKER=${KAFKA_HOST}:9092

ELASTICSEARCH_URL=http://${ELASTICSEARCH_HOST}:9200

# Application Configuration
SERVICE_NAME=$SERVICE_NAME
SERVICE_PORT=$SERVICE_PORT
ENVIRONMENT=production
LOG_LEVEL=info

EOF

chown ubuntu:ubuntu /opt/$SERVICE_NAME/.env

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

# Setup firewall
ufw --force enable
ufw default deny incoming
ufw default allow outgoing
ufw allow 22/tcp
ufw allow $SERVICE_PORT/tcp

# Create systemd service
cat > /etc/systemd/system/$SERVICE_NAME.service << EOF
[Unit]
Description=$SERVICE_NAME
Requires=docker.service
After=docker.service

[Service]
Type=simple
Restart=always
RestartSec=10
User=ubuntu
WorkingDirectory=/opt/$SERVICE_NAME
EnvironmentFile=/opt/$SERVICE_NAME/.env
ExecStart=/usr/bin/docker run --rm --name $SERVICE_NAME \
  --env-file /opt/$SERVICE_NAME/.env \
  -p $SERVICE_PORT:$SERVICE_PORT \
  --network host \
  $SERVICE_NAME:latest
ExecStop=/usr/bin/docker stop $SERVICE_NAME

[Install]
WantedBy=multi-user.target
EOF

# Enable the service
systemctl daemon-reload
systemctl enable $SERVICE_NAME.service

# Install CloudWatch agent for monitoring
wget https://s3.amazonaws.com/amazoncloudwatch-agent/ubuntu/amd64/latest/amazon-cloudwatch-agent.deb
dpkg -i -E ./amazon-cloudwatch-agent.deb
rm amazon-cloudwatch-agent.deb

echo "User data script completed for $SERVICE_NAME!"
