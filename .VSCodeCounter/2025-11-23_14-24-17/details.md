# Details

Date : 2025-11-23 14:24:17

Directory /home/datttne/golang-project

Total : 253 files,  24397 codes, 3304 comments, 5110 blanks, all 32811 lines

[Summary](results.md) / Details / [Diff Summary](diff.md) / [Diff Details](diff-details.md)

## Files
| filename | language | code | comment | blank | total |
| :--- | :--- | ---: | ---: | ---: | ---: |
| [.dockerignore](/.dockerignore) | Ignore | 41 | 17 | 14 | 72 |
| [.github/workflows/ci-cd.yml](/.github/workflows/ci-cd.yml) | YAML | 54 | 0 | 11 | 65 |
| [.github/workflows/generator-generic-ossf-slsa3-publish.yml](/.github/workflows/generator-generic-ossf-slsa3-publish.yml) | YAML | 0 | 32 | 6 | 38 |
| [.github/workflows/go.yml](/.github/workflows/go.yml) | YAML | 0 | 88 | 24 | 112 |
| [AWS-S3-SETUP.md](/AWS-S3-SETUP.md) | Markdown | 100 | 0 | 27 | 127 |
| [CI-CD.md](/CI-CD.md) | Markdown | 59 | 0 | 16 | 75 |
| [DEPLOYMENT.md](/DEPLOYMENT.md) | Markdown | 209 | 0 | 59 | 268 |
| [DOCKER-SETUP.md](/DOCKER-SETUP.md) | Markdown | 111 | 0 | 26 | 137 |
| [README.md](/README.md) | Markdown | 11 | 0 | 2 | 13 |
| [api-gateway/config/config.go](/api-gateway/config/config.go) | Go | 60 | 0 | 9 | 69 |
| [api-gateway/dockerfile](/api-gateway/dockerfile) | Docker | 7 | 4 | 8 | 19 |
| [api-gateway/go.mod](/api-gateway/go.mod) | Go Module File | 46 | 0 | 5 | 51 |
| [api-gateway/go.sum](/api-gateway/go.sum) | Go Checksum File | 124 | 0 | 1 | 125 |
| [api-gateway/helpers/tokenHelpers.go](/api-gateway/helpers/tokenHelpers.go) | Go | 160 | 97 | 49 | 306 |
| [api-gateway/logger/logger.go](/api-gateway/logger/logger.go) | Go | 43 | 0 | 14 | 57 |
| [api-gateway/main.go](/api-gateway/main.go) | Go | 28 | 11 | 11 | 50 |
| [api-gateway/middleware/middleware.go](/api-gateway/middleware/middleware.go) | Go | 214 | 30 | 46 | 290 |
| [api-gateway/models/user\_models.go](/api-gateway/models/user_models.go) | Go | 17 | 0 | 4 | 21 |
| [api-gateway/redisdb/redis.go](/api-gateway/redisdb/redis.go) | Go | 22 | 0 | 9 | 31 |
| [api-gateway/router/routers.go](/api-gateway/router/routers.go) | Go | 501 | 46 | 94 | 641 |
| [auth-service/controller/authController.go](/auth-service/controller/authController.go) | Go | 326 | 8 | 73 | 407 |
| [auth-service/controller/otpController.go](/auth-service/controller/otpController.go) | Go | 122 | 1 | 22 | 145 |
| [auth-service/database/migration/001\_create\_users\_table.down.sql](/auth-service/database/migration/001_create_users_table.down.sql) | MS SQL | 1 | 0 | 0 | 1 |
| [auth-service/database/migration/001\_create\_users\_table.up.sql](/auth-service/database/migration/001_create_users_table.up.sql) | MS SQL | 10 | 0 | 0 | 10 |
| [auth-service/database/migration/002\_create\_valid\_refresh\_token.down.sql](/auth-service/database/migration/002_create_valid_refresh_token.down.sql) | MS SQL | 1 | 0 | 1 | 2 |
| [auth-service/database/migration/002\_create\_valid\_refresh\_token.up.sql](/auth-service/database/migration/002_create_valid_refresh_token.up.sql) | MS SQL | 6 | 0 | 0 | 6 |
| [auth-service/database/postgres.go](/auth-service/database/postgres.go) | Go | 42 | 27 | 17 | 86 |
| [auth-service/database/redis.go](/auth-service/database/redis.go) | Go | 32 | 0 | 8 | 40 |
| [auth-service/dockerfile](/auth-service/dockerfile) | Docker | 7 | 4 | 8 | 19 |
| [auth-service/gRPC/service/user\_service.pb.go](/auth-service/gRPC/service/user_service.pb.go) | Go | 166 | 9 | 28 | 203 |
| [auth-service/gRPC/service/user\_service\_grpc.pb.go](/auth-service/gRPC/service/user_service_grpc.pb.go) | Go | 77 | 29 | 16 | 122 |
| [auth-service/go.mod](/auth-service/go.mod) | Go Module File | 68 | 0 | 6 | 74 |
| [auth-service/go.sum](/auth-service/go.sum) | Go Checksum File | 216 | 0 | 1 | 217 |
| [auth-service/helpers/AuthHelper.go](/auth-service/helpers/AuthHelper.go) | Go | 86 | 21 | 25 | 132 |
| [auth-service/helpers/bloomFilter.go](/auth-service/helpers/bloomFilter.go) | Go | 129 | 1 | 51 | 181 |
| [auth-service/helpers/otp.go](/auth-service/helpers/otp.go) | Go | 56 | 1 | 9 | 66 |
| [auth-service/helpers/redisHelper.go](/auth-service/helpers/redisHelper.go) | Go | 145 | 0 | 61 | 206 |
| [auth-service/helpers/tokenHelper.go](/auth-service/helpers/tokenHelper.go) | Go | 134 | 43 | 35 | 212 |
| [auth-service/kafka/producer.go](/auth-service/kafka/producer.go) | Go | 33 | 1 | 7 | 41 |
| [auth-service/logger/logger.go](/auth-service/logger/logger.go) | Go | 47 | 0 | 14 | 61 |
| [auth-service/main.go](/auth-service/main.go) | Go | 52 | 9 | 21 | 82 |
| [auth-service/middleware/authMiddleware.go](/auth-service/middleware/authMiddleware.go) | Go | 28 | 0 | 3 | 31 |
| [auth-service/models/usersModel.go](/auth-service/models/usersModel.go) | Go | 55 | 29 | 12 | 96 |
| [auth-service/repository/auth-repository.go](/auth-service/repository/auth-repository.go) | Go | 74 | 0 | 16 | 90 |
| [auth-service/routes/AuthRouter.go](/auth-service/routes/AuthRouter.go) | Go | 20 | 0 | 7 | 27 |
| [auth-service/routes/UserRouter.go](/auth-service/routes/UserRouter.go) | Go | 27 | 4 | 7 | 38 |
| [auth-service/service/auth-service.go](/auth-service/service/auth-service.go) | Go | 274 | 22 | 58 | 354 |
| [auth-service/tests/setup\_test.go](/auth-service/tests/setup_test.go) | Go | 15 | 2 | 5 | 22 |
| [auth-service/websocket/websocket.go](/auth-service/websocket/websocket.go) | Go | 71 | 0 | 18 | 89 |
| [auth/googleOAuth.go](/auth/googleOAuth.go) | Go | 67 | 0 | 23 | 90 |
| [backup.sh](/backup.sh) | Shell Script | 60 | 13 | 18 | 91 |
| [cart-service/controller/cartController.go](/cart-service/controller/cartController.go) | Go | 166 | 397 | 107 | 670 |
| [cart-service/controller/server.go](/cart-service/controller/server.go) | Go | 45 | 62 | 32 | 139 |
| [cart-service/dockerfile](/cart-service/dockerfile) | Docker | 7 | 4 | 5 | 16 |
| [cart-service/gRPC/service/cart\_service.pb.go](/cart-service/gRPC/service/cart_service.pb.go) | Go | 216 | 10 | 34 | 260 |
| [cart-service/gRPC/service/cart\_service\_grpc.pb.go](/cart-service/gRPC/service/cart_service_grpc.pb.go) | Go | 77 | 29 | 16 | 122 |
| [cart-service/go.mod](/cart-service/go.mod) | Go Module File | 71 | 0 | 6 | 77 |
| [cart-service/go.sum](/cart-service/go.sum) | Go Checksum File | 215 | 0 | 1 | 216 |
| [cart-service/kafka/consumer.go](/cart-service/kafka/consumer.go) | Go | 55 | 1 | 12 | 68 |
| [cart-service/log/logger.go](/cart-service/log/logger.go) | Go | 44 | 0 | 14 | 58 |
| [cart-service/main.go](/cart-service/main.go) | Go | 111 | 16 | 29 | 156 |
| [cart-service/models/cartModel.go](/cart-service/models/cartModel.go) | Go | 20 | 20 | 8 | 48 |
| [cart-service/repository/cart\_repository.go](/cart-service/repository/cart_repository.go) | Go | 180 | 10 | 42 | 232 |
| [cart-service/routes/cartRouter.go](/cart-service/routes/cartRouter.go) | Go | 21 | 10 | 7 | 38 |
| [cart-service/service/cart-service.go](/cart-service/service/cart-service.go) | Go | 129 | 0 | 37 | 166 |
| [create-topic.sh](/create-topic.sh) | Shell Script | 10 | 3 | 3 | 16 |
| [deploy.sh](/deploy.sh) | Shell Script | 119 | 26 | 31 | 176 |
| [docker-compose.dev.yaml](/docker-compose.dev.yaml) | YAML | 300 | 265 | 44 | 609 |
| [docker-compose.prod.yaml](/docker-compose.prod.yaml) | YAML | 325 | 15 | 18 | 358 |
| [docker-compose.yaml](/docker-compose.yaml) | YAML | 459 | 62 | 29 | 550 |
| [email-service/go.mod](/email-service/go.mod) | Go Module File | 14 | 0 | 6 | 20 |
| [email-service/go.sum](/email-service/go.sum) | Go Checksum File | 77 | 0 | 1 | 78 |
| [email-service/logger/loggger.go](/email-service/logger/loggger.go) | Go | 47 | 0 | 15 | 62 |
| [email-service/main.go](/email-service/main.go) | Go | 17 | 0 | 4 | 21 |
| [email-service/service/email\_service.go](/email-service/service/email_service.go) | Go | 60 | 0 | 10 | 70 |
| [email-service/service/kafka\_consumer.go](/email-service/service/kafka_consumer.go) | Go | 36 | 0 | 9 | 45 |
| [email-service/template/otp\_send.html](/email-service/template/otp_send.html) | HTML | 57 | 0 | 1 | 58 |
| [filebeat.yml](/filebeat.yml) | YAML | 23 | 0 | 4 | 27 |
| [jwt-validator/.traefik.yml](/jwt-validator/.traefik.yml) | YAML | 6 | 0 | 2 | 8 |
| [jwt-validator/go.mod](/jwt-validator/go.mod) | Go Module File | 3 | 0 | 3 | 6 |
| [jwt-validator/go.sum](/jwt-validator/go.sum) | Go Checksum File | 2 | 0 | 1 | 3 |
| [jwt-validator/main.go](/jwt-validator/main.go) | Go | 66 | 1 | 15 | 82 |
| [jwt-validator/vendor/github.com/golang-jwt/jwt/v5/MIGRATION\_GUIDE.md](/jwt-validator/vendor/github.com/golang-jwt/jwt/v5/MIGRATION_GUIDE.md) | Markdown | 157 | 0 | 39 | 196 |
| [jwt-validator/vendor/github.com/golang-jwt/jwt/v5/README.md](/jwt-validator/vendor/github.com/golang-jwt/jwt/v5/README.md) | Markdown | 125 | 0 | 43 | 168 |
| [jwt-validator/vendor/github.com/golang-jwt/jwt/v5/SECURITY.md](/jwt-validator/vendor/github.com/golang-jwt/jwt/v5/SECURITY.md) | Markdown | 10 | 0 | 10 | 20 |
| [jwt-validator/vendor/github.com/golang-jwt/jwt/v5/VERSION\_HISTORY.md](/jwt-validator/vendor/github.com/golang-jwt/jwt/v5/VERSION_HISTORY.md) | Markdown | 97 | 0 | 41 | 138 |
| [jwt-validator/vendor/github.com/golang-jwt/jwt/v5/claims.go](/jwt-validator/vendor/github.com/golang-jwt/jwt/v5/claims.go) | Go | 9 | 6 | 2 | 17 |
| [jwt-validator/vendor/github.com/golang-jwt/jwt/v5/doc.go](/jwt-validator/vendor/github.com/golang-jwt/jwt/v5/doc.go) | Go | 1 | 3 | 1 | 5 |
| [jwt-validator/vendor/github.com/golang-jwt/jwt/v5/ecdsa.go](/jwt-validator/vendor/github.com/golang-jwt/jwt/v5/ecdsa.go) | Go | 92 | 20 | 23 | 135 |
| [jwt-validator/vendor/github.com/golang-jwt/jwt/v5/ecdsa\_utils.go](/jwt-validator/vendor/github.com/golang-jwt/jwt/v5/ecdsa_utils.go) | Go | 51 | 6 | 13 | 70 |
| [jwt-validator/vendor/github.com/golang-jwt/jwt/v5/ed25519.go](/jwt-validator/vendor/github.com/golang-jwt/jwt/v5/ed25519.go) | Go | 52 | 11 | 17 | 80 |
| [jwt-validator/vendor/github.com/golang-jwt/jwt/v5/ed25519\_utils.go](/jwt-validator/vendor/github.com/golang-jwt/jwt/v5/ed25519_utils.go) | Go | 46 | 6 | 13 | 65 |
| [jwt-validator/vendor/github.com/golang-jwt/jwt/v5/errors.go](/jwt-validator/vendor/github.com/golang-jwt/jwt/v5/errors.go) | Go | 59 | 20 | 11 | 90 |
| [jwt-validator/vendor/github.com/golang-jwt/jwt/v5/hmac.go](/jwt-validator/vendor/github.com/golang-jwt/jwt/v5/hmac.go) | Go | 59 | 30 | 16 | 105 |
| [jwt-validator/vendor/github.com/golang-jwt/jwt/v5/map\_claims.go](/jwt-validator/vendor/github.com/golang-jwt/jwt/v5/map_claims.go) | Go | 75 | 16 | 19 | 110 |
| [jwt-validator/vendor/github.com/golang-jwt/jwt/v5/none.go](/jwt-validator/vendor/github.com/golang-jwt/jwt/v5/none.go) | Go | 31 | 8 | 12 | 51 |
| [jwt-validator/vendor/github.com/golang-jwt/jwt/v5/parser.go](/jwt-validator/vendor/github.com/golang-jwt/jwt/v5/parser.go) | Go | 170 | 61 | 38 | 269 |
| [jwt-validator/vendor/github.com/golang-jwt/jwt/v5/parser\_option.go](/jwt-validator/vendor/github.com/golang-jwt/jwt/v5/parser_option.go) | Go | 69 | 61 | 16 | 146 |
| [jwt-validator/vendor/github.com/golang-jwt/jwt/v5/registered\_claims.go](/jwt-validator/vendor/github.com/golang-jwt/jwt/v5/registered_claims.go) | Go | 28 | 22 | 14 | 64 |
| [jwt-validator/vendor/github.com/golang-jwt/jwt/v5/rsa.go](/jwt-validator/vendor/github.com/golang-jwt/jwt/v5/rsa.go) | Go | 62 | 15 | 17 | 94 |
| [jwt-validator/vendor/github.com/golang-jwt/jwt/v5/rsa\_pss.go](/jwt-validator/vendor/github.com/golang-jwt/jwt/v5/rsa_pss.go) | Go | 101 | 16 | 16 | 133 |
| [jwt-validator/vendor/github.com/golang-jwt/jwt/v5/rsa\_utils.go](/jwt-validator/vendor/github.com/golang-jwt/jwt/v5/rsa_utils.go) | Go | 77 | 11 | 20 | 108 |
| [jwt-validator/vendor/github.com/golang-jwt/jwt/v5/signing\_method.go](/jwt-validator/vendor/github.com/golang-jwt/jwt/v5/signing_method.go) | Go | 32 | 8 | 10 | 50 |
| [jwt-validator/vendor/github.com/golang-jwt/jwt/v5/staticcheck.conf](/jwt-validator/vendor/github.com/golang-jwt/jwt/v5/staticcheck.conf) | Properties | 1 | 0 | 1 | 2 |
| [jwt-validator/vendor/github.com/golang-jwt/jwt/v5/token.go](/jwt-validator/vendor/github.com/golang-jwt/jwt/v5/token.go) | Go | 59 | 27 | 15 | 101 |
| [jwt-validator/vendor/github.com/golang-jwt/jwt/v5/token\_option.go](/jwt-validator/vendor/github.com/golang-jwt/jwt/v5/token_option.go) | Go | 2 | 2 | 2 | 6 |
| [jwt-validator/vendor/github.com/golang-jwt/jwt/v5/types.go](/jwt-validator/vendor/github.com/golang-jwt/jwt/v5/types.go) | Go | 80 | 45 | 25 | 150 |
| [jwt-validator/vendor/github.com/golang-jwt/jwt/v5/validator.go](/jwt-validator/vendor/github.com/golang-jwt/jwt/v5/validator.go) | Go | 160 | 121 | 46 | 327 |
| [module/gRPC-Order/go.mod](/module/gRPC-Order/go.mod) | Go Module File | 13 | 0 | 5 | 18 |
| [module/gRPC-Order/go.sum](/module/gRPC-Order/go.sum) | Go Checksum File | 34 | 0 | 1 | 35 |
| [module/gRPC-Order/service/order\_service.pb.go](/module/gRPC-Order/service/order_service.pb.go) | Go | 373 | 13 | 60 | 446 |
| [module/gRPC-Order/service/order\_service\_grpc.pb.go](/module/gRPC-Order/service/order_service_grpc.pb.go) | Go | 149 | 29 | 20 | 198 |
| [module/gRPC-Product/go.mod](/module/gRPC-Product/go.mod) | Go Module File | 13 | 0 | 4 | 17 |
| [module/gRPC-Product/go.sum](/module/gRPC-Product/go.sum) | Go Checksum File | 34 | 0 | 1 | 35 |
| [module/gRPC-Product/service/product\_service.pb.go](/module/gRPC-Product/service/product_service.pb.go) | Go | 726 | 20 | 116 | 862 |
| [module/gRPC-Product/service/product\_service\_grpc.pb.go](/module/gRPC-Product/service/product_service_grpc.pb.go) | Go | 221 | 33 | 24 | 278 |
| [module/gRPC-cart/go.mod](/module/gRPC-cart/go.mod) | Go Module File | 13 | 1 | 6 | 20 |
| [module/gRPC-cart/go.sum](/module/gRPC-cart/go.sum) | Go Checksum File | 34 | 0 | 1 | 35 |
| [module/gRPC-cart/service/cart\_service.pb.go](/module/gRPC-cart/service/cart_service.pb.go) | Go | 214 | 10 | 35 | 259 |
| [module/gRPC-cart/service/cart\_service\_grpc.pb.go](/module/gRPC-cart/service/cart_service_grpc.pb.go) | Go | 77 | 29 | 16 | 122 |
| [order-service/GRPC-PERFORMANCE-FIX.md](/order-service/GRPC-PERFORMANCE-FIX.md) | Markdown | 115 | 0 | 31 | 146 |
| [order-service/controller/order\_controller.go](/order-service/controller/order_controller.go) | Go | 499 | 23 | 103 | 625 |
| [order-service/database/migration/001\_create\_orders\_table.down.sql](/order-service/database/migration/001_create_orders_table.down.sql) | MS SQL | 1 | 0 | 0 | 1 |
| [order-service/database/migration/001\_create\_orders\_table.up.sql](/order-service/database/migration/001_create_orders_table.up.sql) | MS SQL | 15 | 0 | 0 | 15 |
| [order-service/database/postgres.go](/order-service/database/postgres.go) | Go | 39 | 2 | 10 | 51 |
| [order-service/dockerfile](/order-service/dockerfile) | Docker | 13 | 20 | 16 | 49 |
| [order-service/go.mod](/order-service/go.mod) | Go Module File | 63 | 1 | 8 | 72 |
| [order-service/go.sum](/order-service/go.sum) | Go Checksum File | 217 | 0 | 1 | 218 |
| [order-service/kafka/orderProducer.go](/order-service/kafka/orderProducer.go) | Go | 107 | 4 | 26 | 137 |
| [order-service/kafka/paymentProducer.go](/order-service/kafka/paymentProducer.go) | Go | 173 | 4 | 39 | 216 |
| [order-service/kafka/payment\_consumer.go](/order-service/kafka/payment_consumer.go) | Go | 81 | 12 | 22 | 115 |
| [order-service/log/loggger.go](/order-service/log/loggger.go) | Go | 47 | 0 | 15 | 62 |
| [order-service/main.go](/order-service/main.go) | Go | 88 | 100 | 42 | 230 |
| [order-service/models/oderModel.go](/order-service/models/oderModel.go) | Go | 31 | 23 | 8 | 62 |
| [order-service/repositories/order\_repository.go](/order-service/repositories/order_repository.go) | Go | 210 | 15 | 46 | 271 |
| [order-service/routes/order\_route.go](/order-service/routes/order_route.go) | Go | 32 | 1 | 11 | 44 |
| [order-service/service/grpcConnection.go](/order-service/service/grpcConnection.go) | Go | 119 | 11 | 28 | 158 |
| [order-service/service/grpc\_interceptor.go](/order-service/service/grpc_interceptor.go) | Go | 103 | 9 | 20 | 132 |
| [order-service/service/grpc\_service.go](/order-service/service/grpc_service.go) | Go | 57 | 3 | 12 | 72 |
| [order-service/service/health\_check.go](/order-service/service/health_check.go) | Go | 36 | 4 | 14 | 54 |
| [order-service/service/order\_service.go](/order-service/service/order_service.go) | Go | 708 | 53 | 160 | 921 |
| [payment-service/README.md](/payment-service/README.md) | Markdown | 200 | 0 | 52 | 252 |
| [payment-service/database/migration/migration.go](/payment-service/database/migration/migration.go) | Go | 24 | 2 | 7 | 33 |
| [payment-service/database/postgres.go](/payment-service/database/postgres.go) | Go | 14 | 0 | 2 | 16 |
| [payment-service/go.mod](/payment-service/go.mod) | Go Module File | 56 | 0 | 7 | 63 |
| [payment-service/go.sum](/payment-service/go.sum) | Go Checksum File | 304 | 0 | 1 | 305 |
| [payment-service/main.go](/payment-service/main.go) | Go | 39 | 7 | 11 | 57 |
| [payment-service/models/payment\_model.go](/payment-service/models/payment_model.go) | Go | 155 | 20 | 32 | 207 |
| [payment-service/models/vendor\_balance.go](/payment-service/models/vendor_balance.go) | Go | 62 | 6 | 9 | 77 |
| [payment-service/repository/paymentRepository.go](/payment-service/repository/paymentRepository.go) | Go | 195 | 10 | 31 | 236 |
| [payment-service/repository/vendorRepository.go](/payment-service/repository/vendorRepository.go) | Go | 91 | 13 | 18 | 122 |
| [payment-service/routes/route.go](/payment-service/routes/route.go) | Go | 43 | 8 | 12 | 63 |
| [payment-service/src/config/config.go](/payment-service/src/config/config.go) | Go | 18 | 0 | 3 | 21 |
| [payment-service/src/handlers/paymentHandler.go](/payment-service/src/handlers/paymentHandler.go) | Go | 158 | 11 | 35 | 204 |
| [payment-service/src/handlers/vendorHandler.go](/payment-service/src/handlers/vendorHandler.go) | Go | 253 | 21 | 47 | 321 |
| [payment-service/src/main.go](/payment-service/src/main.go) | Go | 100 | 22 | 23 | 145 |
| [payment-service/src/service/bankTransferService.go](/payment-service/src/service/bankTransferService.go) | Go | 130 | 10 | 22 | 162 |
| [payment-service/src/service/grpcConnection.go](/payment-service/src/service/grpcConnection.go) | Go | 15 | 1 | 3 | 19 |
| [payment-service/src/service/kafkaConsumer.go](/payment-service/src/service/kafkaConsumer.go) | Go | 378 | 34 | 83 | 495 |
| [payment-service/src/service/kafkaProducer.go](/payment-service/src/service/kafkaProducer.go) | Go | 144 | 6 | 25 | 175 |
| [payment-service/src/service/paymentGateway.go](/payment-service/src/service/paymentGateway.go) | Go | 574 | 55 | 102 | 731 |
| [payment-service/src/service/refundService.go](/payment-service/src/service/refundService.go) | Go | 41 | 0 | 11 | 52 |
| [payment-service/src/service/vendorService.go](/payment-service/src/service/vendorService.go) | Go | 360 | 35 | 61 | 456 |
| [payment-service/src/utils/logger.go](/payment-service/src/utils/logger.go) | Go | 27 | 0 | 8 | 35 |
| [product-service/config/s3\_config.go](/product-service/config/s3_config.go) | Go | 36 | 0 | 6 | 42 |
| [product-service/controller/productController.go](/product-service/controller/productController.go) | Go | 317 | 73 | 67 | 457 |
| [product-service/controller/server.go](/product-service/controller/server.go) | Go | 107 | 64 | 43 | 214 |
| [product-service/controller/uploadController.go](/product-service/controller/uploadController.go) | Go | 125 | 9 | 24 | 158 |
| [product-service/database/redis.go](/product-service/database/redis.go) | Go | 47 | 6 | 11 | 64 |
| [product-service/dockerfile](/product-service/dockerfile) | Docker | 16 | 23 | 18 | 57 |
| [product-service/go.mod](/product-service/go.mod) | Go Module File | 82 | 2 | 9 | 93 |
| [product-service/go.sum](/product-service/go.sum) | Go Checksum File | 236 | 0 | 1 | 237 |
| [product-service/handler/rating\_update\_handler.go](/product-service/handler/rating_update_handler.go) | Go | 69 | 1 | 12 | 82 |
| [product-service/helper/cached.go](/product-service/helper/cached.go) | Go | 145 | 12 | 34 | 191 |
| [product-service/kafka/consumer.go](/product-service/kafka/consumer.go) | Go | 119 | 2 | 22 | 143 |
| [product-service/kafka/producer.go](/product-service/kafka/producer.go) | Go | 57 | 0 | 20 | 77 |
| [product-service/kafka/rating\_consumer.go](/product-service/kafka/rating_consumer.go) | Go | 68 | 0 | 15 | 83 |
| [product-service/log/logger.go](/product-service/log/logger.go) | Go | 47 | 0 | 14 | 61 |
| [product-service/main.go](/product-service/main.go) | Go | 157 | 13 | 41 | 211 |
| [product-service/models/productModel.go](/product-service/models/productModel.go) | Go | 82 | 100 | 25 | 207 |
| [product-service/repository/product-repository.go](/product-service/repository/product-repository.go) | Go | 424 | 46 | 75 | 545 |
| [product-service/routes/productManagerRouter.go](/product-service/routes/productManagerRouter.go) | Go | 45 | 9 | 15 | 69 |
| [product-service/routes/productUploadRoutes.go](/product-service/routes/productUploadRoutes.go) | Go | 13 | 11 | 6 | 30 |
| [product-service/routes/uploadRoutes.go](/product-service/routes/uploadRoutes.go) | Go | 13 | 2 | 5 | 20 |
| [product-service/s3/s3Client.go](/product-service/s3/s3Client.go) | Go | 18 | 0 | 4 | 22 |
| [product-service/service/product\_service.go](/product-service/service/product_service.go) | Go | 320 | 13 | 45 | 378 |
| [product-service/service/s3\_service.go](/product-service/service/s3_service.go) | Go | 191 | 19 | 50 | 260 |
| [product-service/test/test.go](/product-service/test/test.go) | Go | 225 | 43 | 52 | 320 |
| [restore.sh](/restore.sh) | Shell Script | 107 | 12 | 18 | 137 |
| [review-service/README.md](/review-service/README.md) | Markdown | 58 | 0 | 15 | 73 |
| [review-service/cmd/server/main.go](/review-service/cmd/server/main.go) | Go | 81 | 6 | 23 | 110 |
| [review-service/config/config.go](/review-service/config/config.go) | Go | 60 | 0 | 9 | 69 |
| [review-service/config/s3Config.go](/review-service/config/s3Config.go) | Go | 36 | 0 | 6 | 42 |
| [review-service/go.mod](/review-service/go.mod) | Go Module File | 65 | 0 | 7 | 72 |
| [review-service/go.sum](/review-service/go.sum) | Go Checksum File | 167 | 0 | 1 | 168 |
| [review-service/internal/cron/pending\_review\_aggregator.go](/review-service/internal/cron/pending_review_aggregator.go) | Go | 156 | 10 | 34 | 200 |
| [review-service/internal/cron/scheduler.go](/review-service/internal/cron/scheduler.go) | Go | 33 | 4 | 10 | 47 |
| [review-service/internal/handlers/review\_handler.go](/review-service/internal/handlers/review_handler.go) | Go | 99 | 1 | 19 | 119 |
| [review-service/internal/kafka/producer.go](/review-service/internal/kafka/producer.go) | Go | 72 | 0 | 17 | 89 |
| [review-service/internal/models/review.go](/review-service/internal/models/review.go) | Go | 17 | 0 | 4 | 21 |
| [review-service/internal/models/review\_aggregate.go](/review-service/internal/models/review_aggregate.go) | Go | 20 | 0 | 4 | 24 |
| [review-service/internal/repository/review\_repository.go](/review-service/internal/repository/review_repository.go) | Go | 154 | 1 | 28 | 183 |
| [review-service/internal/routes/routes.go](/review-service/internal/routes/routes.go) | Go | 20 | 1 | 6 | 27 |
| [review-service/internal/services/review\_services.go](/review-service/internal/services/review_services.go) | Go | 47 | 0 | 10 | 57 |
| [review-service/log/logger.go](/review-service/log/logger.go) | Go | 47 | 0 | 14 | 61 |
| [scripts/push-ecr-images.sh](/scripts/push-ecr-images.sh) | Shell Script | 60 | 10 | 16 | 86 |
| [scripts/remote-deploy.sh](/scripts/remote-deploy.sh) | Shell Script | 36 | 10 | 11 | 57 |
| [search-service/controller/search\_controller.go](/search-service/controller/search_controller.go) | Go | 60 | 0 | 8 | 68 |
| [search-service/database/elasticsearch.go](/search-service/database/elasticsearch.go) | Go | 18 | 0 | 9 | 27 |
| [search-service/dockerfile](/search-service/dockerfile) | Docker | 8 | 0 | 5 | 13 |
| [search-service/go.mod](/search-service/go.mod) | Go Module File | 54 | 2 | 9 | 65 |
| [search-service/go.sum](/search-service/go.sum) | Go Checksum File | 183 | 0 | 1 | 184 |
| [search-service/kafka/consumer.go](/search-service/kafka/consumer.go) | Go | 74 | 3 | 14 | 91 |
| [search-service/log/logger.go](/search-service/log/logger.go) | Go | 47 | 0 | 14 | 61 |
| [search-service/main.go](/search-service/main.go) | Go | 52 | 0 | 20 | 72 |
| [search-service/models/product.go](/search-service/models/product.go) | Go | 9 | 0 | 1 | 10 |
| [search-service/repository/search\_repository.go](/search-service/repository/search_repository.go) | Go | 194 | 2 | 41 | 237 |
| [search-service/routes/search\_router.go](/search-service/routes/search_router.go) | Go | 9 | 0 | 2 | 11 |
| [search-service/service/search\_service.go](/search-service/service/search_service.go) | Go | 65 | 0 | 14 | 79 |
| [start-dev.ps1](/start-dev.ps1) | PowerShell | 20 | 3 | 4 | 27 |
| [start-prod.ps1](/start-prod.ps1) | PowerShell | 20 | 3 | 4 | 27 |
| [stop-all.ps1](/stop-all.ps1) | PowerShell | 8 | 4 | 5 | 17 |
| [terraform/MULTI-INSTANCE-DEPLOYMENT.md](/terraform/MULTI-INSTANCE-DEPLOYMENT.md) | Markdown | 323 | 0 | 86 | 409 |
| [terraform/README.md](/terraform/README.md) | Markdown | 323 | 0 | 132 | 455 |
| [terraform/infrastructure\_user\_data.sh](/terraform/infrastructure_user_data.sh) | Shell Script | 100 | 17 | 19 | 136 |
| [terraform/service\_user\_data.sh](/terraform/service_user_data.sh) | Shell Script | 97 | 17 | 27 | 141 |
| [terraform/shared\_infrastructure\_user\_data.sh](/terraform/shared_infrastructure_user_data.sh) | Shell Script | 149 | 14 | 24 | 187 |
| [terraform/traefik\_user\_data.sh](/terraform/traefik_user_data.sh) | Shell Script | 148 | 12 | 34 | 194 |
| [terraform/user\_data.sh](/terraform/user_data.sh) | Shell Script | 108 | 22 | 36 | 166 |
| [traefik-config/dynamic.yml](/traefik-config/dynamic.yml) | YAML | 35 | 0 | 3 | 38 |
| [traefik-config/traefik.yml](/traefik-config/traefik.yml) | YAML | 4 | 0 | 0 | 4 |
| [user-service/Makefile](/user-service/Makefile) | Makefile | 12 | 1 | 6 | 19 |
| [user-service/README.md](/user-service/README.md) | Markdown | 81 | 0 | 20 | 101 |
| [user-service/cmd/server/main.go](/user-service/cmd/server/main.go) | Go | 53 | 5 | 14 | 72 |
| [user-service/configs/config.yaml](/user-service/configs/config.yaml) | YAML | 28 | 0 | 6 | 34 |
| [user-service/go.mod](/user-service/go.mod) | Go Module File | 55 | 0 | 6 | 61 |
| [user-service/go.sum](/user-service/go.sum) | Go Checksum File | 117 | 0 | 1 | 118 |
| [user-service/internal/config/config.go](/user-service/internal/config/config.go) | Go | 55 | 4 | 11 | 70 |
| [user-service/internal/events/consumer.go](/user-service/internal/events/consumer.go) | Go | 64 | 4 | 9 | 77 |
| [user-service/internal/events/kafka\_publisher.go](/user-service/internal/events/kafka_publisher.go) | Go | 38 | 1 | 10 | 49 |
| [user-service/internal/events/publisher.go](/user-service/internal/events/publisher.go) | Go | 18 | 2 | 6 | 26 |
| [user-service/internal/handlers/users.go](/user-service/internal/handlers/users.go) | Go | 135 | 0 | 25 | 160 |
| [user-service/internal/models/users\_model.go](/user-service/internal/models/users_model.go) | Go | 30 | 0 | 5 | 35 |
| [user-service/internal/repository/users\_repo.go](/user-service/internal/repository/users_repo.go) | Go | 128 | 0 | 26 | 154 |
| [user-service/internal/routes/routes.go](/user-service/internal/routes/routes.go) | Go | 21 | 0 | 5 | 26 |
| [user-service/internal/services/users\_service.go](/user-service/internal/services/users_service.go) | Go | 59 | 1 | 14 | 74 |
| [user-service/log/logger.go](/user-service/log/logger.go) | Go | 47 | 0 | 14 | 61 |
| [user-service/migrations/0001\_create\_users\_table.down.sql](/user-service/migrations/0001_create_users_table.down.sql) | MS SQL | 3 | 0 | 0 | 3 |
| [user-service/migrations/0001\_create\_users\_table.up.sql](/user-service/migrations/0001_create_users_table.up.sql) | MS SQL | 27 | 3 | 4 | 34 |
| [user-service/migrations/0002\_add\_default\_address\_id\_column.down.sql](/user-service/migrations/0002_add_default_address_id_column.down.sql) | MS SQL | 2 | 0 | 0 | 2 |
| [user-service/migrations/0002\_add\_default\_address\_id\_column.up.sql](/user-service/migrations/0002_add_default_address_id_column.up.sql) | MS SQL | 2 | 0 | 0 | 2 |
| [user-service/scripts/migrate.sh](/user-service/scripts/migrate.sh) | Shell Script | 3 | 3 | 3 | 9 |
| [validate-env.ps1](/validate-env.ps1) | PowerShell | 44 | 2 | 10 | 56 |

[Summary](results.md) / Details / [Diff Summary](diff.md) / [Diff Details](diff-details.md)