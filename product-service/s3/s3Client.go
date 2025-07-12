package s3

import (
	"product-service/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)


func NewS3Client(cfg *config.S3Config) *s3.S3 {
	awsCfg := &aws.Config{
		Region:      aws.String(cfg.Region),
	}

	if cfg.Endpoint != "" {
		awsCfg.Endpoint = aws.String(cfg.Endpoint)
		awsCfg.S3ForcePathStyle = aws.Bool(true) // For localstack or custom endpoints
	}
	sess := session.Must(session.NewSession(awsCfg))
	return s3.New(sess)
}