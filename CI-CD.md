# CI/CD Guide - Build, Push, and Deploy

This guide covers how to use the GitHub Actions workflow to build Docker images, push them to ECR, and deploy to EC2 instances.

## What the workflow does
- Builds Docker images for all microservices
- Pushes them to ECR
- Deploys to the service instances by SSHing into each server and running `remote-deploy.sh`

## Required GitHub Secrets
- `AWS_ACCOUNT_ID` - your AWS Account ID
- `AWS_REGION` - region (e.g. us-east-1)
- `AWS_ROLE_TO_ASSUME` - role ARN to assume via OIDC for ECR push (recommended)
- `SSH_USER` - SSH username (e.g. `ubuntu`)
- `SSH_PRIVATE_KEY` - Private key content for SSH (use for Appleboy's action)

Per-service host secrets (one per service, example):
- `HOST_API_GATEWAY` - IP or hostname for API Gateway (Traefik server)
- `HOST_AUTH_SERVICE` - IP for auth-service instance
- `HOST_USER_SERVICE` - IP for user-service instance
- ...etc. Use uppercase and replace hyphens with underscores.

## Files
- `scripts/push-ecr-images.sh` - Build and push Docker images to ECR
- `scripts/remote-deploy.sh` - Remote script to pull image and restart the systemd service
- `.github/workflows/ci-cd.yml` - GitHub Actions workflow

## How to use
- Add the required secrets to GitHub repository
- Ensure `remote-deploy.sh` is placed under `/opt/microservices/` on EC2 instances and marked executable
- Ensure `systemd` units exist and use container name same as service and image pointing to ECR tag
- Push to `main` branch to trigger the workflow or run manually

## Security notes
- Use AWS OIDC with a role that has limited permissions to push images to ECR
- Use an SSH key specifically for CI and limit source IPs where possible
- Do not store private keys in code; use GitHub Secrets

## Troubleshooting
- If ECR push fails: verify the role permissions and the repo region
- If SSH step fails: verify host secret and key, ensure SSH access allowed in security group, and the `remote-deploy.sh` exists
- If systemd doesn't pick up the new image: ensure service Unit doesn't pin a specific tag, or add a unit command to pull latest image as part of the unit file

## Next steps (optional enhancements)
- Add a rolling update mechanism
- Use SSM (AWS Systems Manager) instead of SSH for deployment
- Use autoscaling groups + AMI baking for large horizontal scaling
- Integrate secrets management (HashiCorp Vault / AWS Secrets Manager)

# Example: Adding repo secrets
1. Open GitHub -> Settings -> Secrets -> Actions
2. Add `AWS_ACCOUNT_ID`, `AWS_REGION`, `AWS_ROLE_TO_ASSUME`, `SSH_USER`, `SSH_PRIVATE_KEY` and per-service host secrets like `HOST_AUTH_SERVICE`

# Example systemd Unit
A simple unit for `auth-service` that pulls a specific image tag may look like:

```
[Unit]
Description=auth-service
Requires=docker.service
After=docker.service

[Service]
Restart=always
User=ubuntu
ExecStartPre=-/usr/bin/docker pull <account>.dkr.ecr.us-east-1.amazonaws.com/auth-service:latest
ExecStart=/usr/bin/docker run --name auth-service --rm -p 8081:8081 --env-file /opt/auth-service/.env <account>.dkr.ecr.us-east-1.amazonaws.com/auth-service:latest
ExecStop=/usr/bin/docker stop auth-service

[Install]
WantedBy=multi-user.target
```

Note: The `remote-deploy.sh` will pull the new image and start the unit.
