module product-service

go 1.24.4

require (
	// github.com/Dattt2k2/golang-project v0.0.0-20250628100049-f9c5be5cbd05
	// github.com/Dattt2k2/golang-project/auth-service v0.0.0-00010101000000-000000000000
	github.com/aws/aws-sdk-go v1.55.7
	github.com/gin-gonic/gin v1.10.1
	github.com/joho/godotenv v1.5.1
	github.com/redis/go-redis/v9 v9.11.0
	github.com/segmentio/kafka-go v0.4.48
	github.com/stretchr/testify v1.10.0
	go.mongodb.org/mongo-driver v1.17.4
	go.uber.org/zap v1.27.0
	google.golang.org/grpc v1.73.0
)

require module/gRPC-Product v0.0.0-00010101000000-000000000000

require (
	github.com/bytedance/sonic v1.11.6 // indirect
	github.com/bytedance/sonic/loader v0.1.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cloudwego/base64x v0.1.4 // indirect
	github.com/cloudwego/iasm v0.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.23.0 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/klauspost/cpuid/v2 v2.2.7 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/arch v0.8.0 // indirect
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sync v0.15.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250324211829-b45e905df463 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/Dattt2k2/golang-project/module/gRPC-Product => ../module/gRPC-Product

replace github.com/Dattt2k2/golang-project/module/gRPC-cart => ../module/gRPC-cart

replace module/gRPC-Product => ../module/gRPC-Product
