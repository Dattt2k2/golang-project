#!/bin/bash
set -e

# Infrastructure service configuration
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
    htop \
    iotop \
    nethogs

# Install Docker
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
apt-get update
apt-get install -y docker-ce docker-ce-cli containerd.io

# Start and enable Docker
systemctl start docker
systemctl enable docker

# Create docker group
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

# Create data directory
mkdir -p /data/$SERVICE_NAME
chown ubuntu:ubuntu /data/$SERVICE_NAME

# Setup service based on type
case $SERVICE_NAME in
  redis)
    # Run Redis
    docker run -d \
      --name redis \
      --restart unless-stopped \
      -p 6379:6379 \
      -v /data/redis:/data \
      redis:7-alpine \
      redis-server --appendonly yes
    ;;
    
  kafka)
    # Run Kafka in KRaft mode
    docker run -d \
      --name kafka \
      --restart unless-stopped \
      -p 9092:9092 \
      -p 9093:9093 \
      -v /data/kafka:/var/lib/kafka/data \
      -e KAFKA_NODE_ID=1 \
      -e KAFKA_PROCESS_ROLES=broker,controller \
      -e KAFKA_LISTENERS=PLAINTEXT://0.0.0.0:9092,CONTROLLER://0.0.0.0:9093 \
      -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://$(hostname -I | awk '{print $1}'):9092 \
      -e KAFKA_CONTROLLER_LISTENER_NAMES=CONTROLLER \
      -e KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT \
      -e KAFKA_CONTROLLER_QUORUM_VOTERS=1@$(hostname -I | awk '{print $1}'):9093 \
      -e KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1 \
      -e KAFKA_TRANSACTION_STATE_LOG_MIN_ISR=1 \
      -e KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR=1 \
      -e KAFKA_AUTO_CREATE_TOPICS_ENABLE=true \
      -e KAFKA_LOG_DIRS=/var/lib/kafka/data \
      -e CLUSTER_ID=MkU3OEVBNTcwNTJENDM2Qk \
      confluentinc/cp-kafka:7.5.0
    ;;
    
  elasticsearch)
    # Run Elasticsearch
    docker run -d \
      --name elasticsearch \
      --restart unless-stopped \
      -p 9200:9200 \
      -v /data/elasticsearch:/usr/share/elasticsearch/data \
      -e "discovery.type=single-node" \
      -e "xpack.security.enabled=false" \
      -e "ES_JAVA_OPTS=-Xms2g -Xmx2g" \
      docker.elastic.co/elasticsearch/elasticsearch:8.11.0
    
    # Run Kibana
    sleep 30
    docker run -d \
      --name kibana \
      --restart unless-stopped \
      -p 5601:5601 \
      -e "ELASTICSEARCH_HOSTS=http://localhost:9200" \
      docker.elastic.co/kibana/kibana:8.11.0
    ;;
esac

# Setup firewall
ufw --force enable
ufw default deny incoming
ufw default allow outgoing
ufw allow 22/tcp
ufw allow $SERVICE_PORT/tcp

# Additional ports for specific services
if [ "$SERVICE_NAME" = "kafka" ]; then
  ufw allow 9093/tcp
fi

if [ "$SERVICE_NAME" = "elasticsearch" ]; then
  ufw allow 5601/tcp
fi

# Install CloudWatch agent
wget https://s3.amazonaws.com/amazoncloudwatch-agent/ubuntu/amd64/latest/amazon-cloudwatch-agent.deb
dpkg -i -E ./amazon-cloudwatch-agent.deb
rm amazon-cloudwatch-agent.deb

echo "Infrastructure service $SERVICE_NAME started successfully!"
