module email-service

go 1.24.4

require go.uber.org/zap v1.27.0

require (
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/stretchr/testify v1.8.3 // indirect
)

require (
	github.com/joho/godotenv v1.5.1
	github.com/segmentio/kafka-go v0.4.48
	go.uber.org/multierr v1.10.0 // indirect
)

replace github.com/Dattt2k2/golang-project/email-service => ../email-service
