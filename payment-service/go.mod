module payment-service

go 1.24.4

require (
	github.com/segmentio/kafka-go v0.4.48
	google.golang.org/grpc v1.51.0
)

require (
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	google.golang.org/genproto v0.0.0-20220503193339-ba3ae3f07e29 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
)

replace github.com/Dattt2k2/golang-project/payment-service => ../payment-service

replace github.com/Dattt2k2/golang-project/module/gRPC-Order => ../module/gRPC-Order/service
