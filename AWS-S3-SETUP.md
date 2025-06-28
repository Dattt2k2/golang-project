# AWS S3 Setup Guide

## 📋 **Prerequisites**

1. **AWS Account** với S3 access
2. **IAM User** với S3 permissions
3. **S3 Bucket** đã tạo sẵn

## 🔧 **AWS Setup Steps**

### 1. Tạo S3 Bucket
```bash
# Tạo bucket (replace your-bucket-name)
aws s3 mb s3://your-bucket-name --region ap-southeast-1

# Set bucket policy for public read (optional)
aws s3api put-bucket-policy --bucket your-bucket-name --policy file://bucket-policy.json
```

### 2. Tạo IAM User
- Vào AWS Console → IAM → Users → Create User
- Attach policy: `AmazonS3FullAccess` hoặc custom policy
- Tạo Access Key và lưu lại

### 3. Example IAM Policy (minimal permissions)
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:PutObject",
                "s3:GetObject",
                "s3:DeleteObject"
            ],
            "Resource": "arn:aws:s3:::your-bucket-name/*"
        }
    ]
}
```

## 🛠️ **Project Setup**

### 1. Cập nhật .env file
```bash
# Copy template
cp .env.s3.template product-service/.env.s3

# Edit product-service/.env
AWS_REGION=ap-southeast-1
AWS_ACCESS_KEY_ID=AKIA...your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-access-key
AWS_S3_BUCKET=your-bucket-name
AWS_S3_FOLDER=product-images
```

### 2. Install dependencies
```bash
# Trong container hoặc local
go mod tidy
```

### 3. Test upload
```bash
# Start development environment
.\start-dev.ps1

# Test upload endpoint
curl -X POST http://localhost:8080/products/upload \
  -F "file=@test-image.jpg" \
  -H "Authorization: Bearer your-jwt-token"
```

## 📁 **File Structure**

```
product-service/
├── .env                    # Contains AWS credentials
├── config/
│   └── s3_config.go       # S3 configuration loader
├── service/
│   └── s3_service.go      # S3 upload/delete service
└── controller/
    └── upload_controller.go # HTTP handlers (create this)
```

## 🔐 **Security Best Practices**

1. **Never commit .env files**
```bash
echo "*.env" >> .gitignore
```

2. **Use IAM roles** trong production thay vì hardcode credentials

3. **Validate file types** và sizes trước khi upload

4. **Set proper CORS** policy cho S3 bucket nếu cần

## 🚀 **Usage in Docker**

Environment variables từ `.env` file sẽ tự động được load vào container:

```yaml
# docker-compose.dev.yaml
product-service:
  env_file:
    - ./product-service/.env  # AWS credentials loaded here
  environment:
    - REDIS_URL=redis:6379    # Override specific vars if needed
```

## 🔍 **Troubleshooting**

### Common Issues:
1. **403 Forbidden**: Check IAM permissions
2. **InvalidAccessKeyId**: Verify credentials trong .env
3. **NoSuchBucket**: Verify bucket name và region
4. **File too large**: Check MAX_FILE_SIZE setting

### Debug trong container:
```bash
docker exec -it golang-project-product-service-1 sh
printenv | grep AWS
```
