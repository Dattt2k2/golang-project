done
#!/usr/bin/env bash
set -euo pipefail

# Build and push Docker images for microservices to ECR
# Usage: ./push-ecr-images.sh <aws_account_id> <aws_region> [tag]

AWS_ACCOUNT_ID="${1:-}"
AWS_REGION="${2:-us-east-1}"
TAG="${3:-${GITHUB_SHA:-latest}}"
ECR_REGISTRY="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com"

if [ -z "${AWS_ACCOUNT_ID}" ]; then
  echo "Usage: $0 <aws_account_id> <aws_region> [tag]"
  exit 1
fi

# Tools
for cmd in aws docker jq; do
  command -v ${cmd} >/dev/null 2>&1 || { echo "${cmd} is required but not installed"; exit 1; }
done

SERVICES=(api-gateway auth-service user-service product-service cart-service order-service payment-service search-service review-service email-service)

echo "ECR Registry: ${ECR_REGISTRY}"
echo "Using tag: ${TAG}"

# Ensure ECR repos exist (idempotent)
for svc in "${SERVICES[@]}"; do
  if ! aws ecr describe-repositories --repository-names "${svc}" --region "${AWS_REGION}" >/dev/null 2>&1; then
    aws ecr create-repository --repository-name "${svc}" --region "${AWS_REGION}" >/dev/null
    echo "ECR: Created ${svc}"
  fi
done

# Login
aws ecr get-login-password --region "${AWS_REGION}" | docker login --username AWS --password-stdin "${ECR_REGISTRY}"

# Build/push
for svc in "${SERVICES[@]}"; do
  echo "\n== Building: ${svc} =="

  # locate context and Dockerfile
  CONTEXT_DIR="${svc}"
  if [ ! -d "${CONTEXT_DIR}" ]; then
    # fallback try different cases
    if [ -d "${svc^}" ]; then
      CONTEXT_DIR="${svc^}"
    elif [ -d "${svc,,}" ]; then
      CONTEXT_DIR="${svc,,}"
    fi
  fi
  if [ ! -d "${CONTEXT_DIR}" ]; then
    echo "Skipping ${svc}: directory not found"
    continue
  fi

  DOCKERFILE="${CONTEXT_DIR}/Dockerfile"
  if [ ! -f "${DOCKERFILE}" ]; then
    if [ -f "${CONTEXT_DIR}/dockerfile" ]; then
      DOCKERFILE="${CONTEXT_DIR}/dockerfile"
    else
      echo "Skipping ${svc}: Dockerfile not found in ${CONTEXT_DIR}"
      continue
    fi
  fi

  IMAGE_URI="${ECR_REGISTRY}/${svc}:${TAG}"
  LATEST_URI="${ECR_REGISTRY}/${svc}:latest"

  # build using docker buildx for better caching; fallback to plain docker
  if docker buildx version >/dev/null 2>&1; then
    docker buildx build --platform linux/amd64 -t "${IMAGE_URI}" -f "${DOCKERFILE}" --push "${CONTEXT_DIR}"
  else
    docker build -t "${IMAGE_URI}" -f "${DOCKERFILE}" "${CONTEXT_DIR}"
    docker push "${IMAGE_URI}"
  fi

  echo "Tagging latest: ${LATEST_URI}"
  docker tag "${IMAGE_URI}" "${LATEST_URI}" || true
  docker push "${LATEST_URI}" || true

  echo "Pushed ${svc} -> ${IMAGE_URI}"
done

echo "All images pushed successfully."