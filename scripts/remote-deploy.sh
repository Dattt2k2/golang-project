#!/usr/bin/env bash
set -euo pipefail

# Remote deploy script to run on EC2 instance
# Usage: ./remote-deploy.sh <service> <aws_account_id> <aws_region> <tag>

SERVICE="${1:-}"
AWS_ACCOUNT_ID="${2:-}"
AWS_REGION="${3:-us-east-1}"
TAG="${4:-latest}"

if [ -z "${SERVICE}" ] || [ -z "${AWS_ACCOUNT_ID}" ]; then
  echo "Usage: $0 <service> <aws_account_id> <aws_region> <tag>"
  exit 1
fi

ECR_REGISTRY="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com"
IMAGE_URI="${ECR_REGISTRY}/${SERVICE}:${TAG}"

# Login
aws ecr get-login-password --region "${AWS_REGION}" | docker login --username AWS --password-stdin "${ECR_REGISTRY}"

# Pull latest image
echo "Pulling ${IMAGE_URI}..."
docker pull "${IMAGE_URI}"

# If using systemd service with name "<service>.service" and runs the container name equal to service
CONTAINER_NAME="${SERVICE}"
SERVICE_UNIT="${SERVICE}.service"

# Stop existing container (if systemd)
if systemctl is-active --quiet "${SERVICE_UNIT}"; then
  echo "Stopping systemd service ${SERVICE_UNIT}..."
  systemctl stop "${SERVICE_UNIT}"
fi

# Remove old container (if exists)
if docker ps -a --format "{{.Names}}" | grep -q "^${CONTAINER_NAME}$"; then
  echo "Removing old container ${CONTAINER_NAME}..."
  docker rm -f "${CONTAINER_NAME}" || true
fi

# Start new container (will rely on systemd to start with updated image) or run directly
# Systemd unit should use the container image name tag, so we `docker pull` and then start the unit
if systemctl list-unit-files | grep -q "${SERVICE_UNIT}"; then
  echo "Starting systemd service ${SERVICE_UNIT}..."
  systemctl daemon-reload || true
  systemctl start "${SERVICE_UNIT}"
else
  echo "No systemd unit found for ${SERVICE}, running container directly..."
  docker run -d --name "${CONTAINER_NAME}" --restart unless-stopped \
    -e ENVIRONMENT=production \
    -p ${PORT:-8080}:${PORT:-8080} \
    "${IMAGE_URI}"
fi

echo "Deployed ${SERVICE} successfully."