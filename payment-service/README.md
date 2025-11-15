# Payment Service

This project is a payment service that integrates with Stripe payment gateway and uses Kafka for communication with other services. It is designed to handle payment processing, manage payment-related events, and support Stripe Connect for multi-vendor payments.

## Features

- **Stripe Payment Integration**: Full support for PaymentIntent API with manual capture (escrow)
- **Stripe Connect**: Multi-vendor payment splitting with platform fees
- **Webhook Security**: HMAC signature verification for internal webhooks and Stripe signature validation
- **Event-Driven**: Kafka integration for async communication
- **Database**: PostgreSQL for payment records
- **Refund Support**: Full refund processing capabilities

## Project Structure

```
payment-service
├── src
│   ├── main.go                # Entry point of the application
│   ├── service                # Contains service-related logic
│   │   ├── paymentGateway.go   # Stripe payment gateway integration
│   │   ├── refundService.go    # Refund processing
│   │   ├── vendorService.go    # Stripe Connect vendor management
│   │   ├── kafkaProducer.go    # Kafka producer implementation
│   │   ├── kafkaConsumer.go    # Kafka consumer implementation
│   │   └── grpcConnection.go   # gRPC connection to order service
│   ├── config                 # Configuration settings
│   │   └── config.go          # Configuration struct and settings
│   ├── handlers               # HTTP request handlers
│   │   ├── paymentHandler.go   # Payment request handling
│   │   └── vendorHandler.go    # Vendor/Connect handling
│   ├── models                 # Data structures
│   │   └── payment.go         # Payment-related structs
│   └── utils                  # Utility functions
│       └── logger.go          # Logging utilities
├── routes                     # API routes
│   └── route.go               # Route definitions
├── repository                 # Database layer
│   ├── payment_repository.go  # Payment database operations
│   └── vendor_repository.go   # Vendor database operations
├── database                   # Database connection
│   └── postgres.go            # PostgreSQL connection
├── go.mod                     # Module definition and dependencies
├── go.sum                     # Dependency checksums
└── README.md                  # Project documentation
```

## Environment Variables

```bash
# Stripe Configuration
STRIPE_API_KEY=sk_test_xxxxx                    # Stripe secret key
STRIPE_WEBHOOK_SECRET=whsec_xxxxx               # Stripe webhook signing secret

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=payment_db

# Kafka
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC_PAYMENT=payment_events

# Service
PORT=8088
```

## Setup Instructions

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd payment-service
   ```

2. **Install dependencies:**
   ```bash
   go mod tidy
   ```

3. **Configure environment variables:**
   Create a `.env` file or set environment variables as shown above.

4. **Setup Stripe Webhook:**
   ```bash
   # For local development, use Stripe CLI
   stripe listen --forward-to localhost:8088/webhook/stripe
   
   # This will provide a webhook signing secret (whsec_xxx)
   # Add it to your .env as STRIPE_WEBHOOK_SECRET
   ```

5. **Run the service:**
   ```bash
   go run main.go
   ```

## API Endpoints

### Payment Routes
- `POST /api/v1/payments` - Create a new payment
- `GET /api/v1/payments/:order_id` - Get payment by order ID

### Refund Routes
- `POST /api/v1/refunds` - Process a refund
- `GET /api/v1/refunds/:refund_id` - Get refund details

### Vendor Routes (Stripe Connect)
- `POST /api/v1/vendors/register` - Register a new vendor
- `GET /api/v1/vendors/:vendor_id` - Get vendor details
- `POST /api/v1/vendors/:vendor_id/onboarding` - Create onboarding link
- `GET /api/v1/vendors/:vendor_id/onboarding/status` - Check onboarding status

### Webhook Routes
- `POST /webhook/stripe` - **Official Stripe webhook endpoint** (use this in Stripe Dashboard)
- `POST /webhook/payment` - Internal payment webhook (for internal services)
- `POST /webhook/refund` - Internal refund webhook
- `POST /webhook/stripe/connect` - Stripe Connect events

### Public Routes
- `GET /public/vendor/onboarding/success` - Vendor onboarding success redirect
- `GET /public/vendor/onboarding/refresh` - Vendor onboarding refresh redirect

## Webhook Configuration

### Stripe Webhook Setup

1. **In Stripe Dashboard:**
   - Go to Developers → Webhooks
   - Click "Add endpoint"
   - URL: `https://yourdomain.com/webhook/stripe`
   - Select events to listen for:
     - `payment_intent.succeeded`
     - `payment_intent.captured`
     - `payment_intent.payment_failed`
     - `transfer.created`
     - `transfer.paid`
     - `transfer.failed`
   - Copy the webhook signing secret

2. **For Local Development:**
   ```bash
   stripe listen --forward-to localhost:8088/webhook/stripe
   ```

3. **Security:**
   - Webhook signature is automatically verified using `Stripe-Signature` header
   - All webhook events are logged for debugging
   - Invalid signatures return 400 Bad Request

### Internal Webhook (Optional)

For internal service-to-service communication:
- Endpoint: `/webhook/payment`
- Header: `X-Signature` with HMAC-SHA256 signature
- Used for custom payment status updates from other services

## Payment Flow

### 1. Create Payment (Hold/Escrow)
```json
POST /api/v1/payments
{
  "order_id": "order_123",
  "amount": 100.00,
  "currency": "usd"
}
```

Response includes `client_secret` for frontend to confirm payment method.

### 2. Stripe Webhook Events

- **payment_intent.succeeded**: Payment method authorized (money held)
- **payment_intent.captured**: Payment captured (money transferred)
- **payment_intent.payment_failed**: Payment failed

### 3. Capture Payment

After order is confirmed/shipped:
```go
// Capture is done via Stripe API
// Webhook will notify service when captured
```

## Multi-Vendor Payments (Stripe Connect)

### 1. Register Vendor
```json
POST /api/v1/vendors/register
{
  "vendor_id": "vendor_123",
  "email": "vendor@example.com",
  "country": "US"
}
```

### 2. Create Payment with Vendor Split
```json
POST /api/v1/payments
{
  "order_id": "order_123",
  "amount": 100.00,
  "currency": "usd",
  "vendor_stripe_account_id": "acct_xxx",
  "platform_fee": 5.00,
  "vendor_breakdown": "vendor1:50,vendor2:45"
}
```

### 3. Stripe Transfers
- Platform receives full payment
- Stripe automatically transfers vendor portion
- Webhook events: `transfer.created`, `transfer.paid`, `transfer.failed`

## Logging

All webhook events are logged with format:
```
[Webhook] Received event: payment_intent.succeeded (ID: evt_xxx)
[Webhook] Payment succeeded for order: order_123, PaymentIntent: pi_xxx
```

## Testing

### Test Stripe Webhook Locally
```bash
# Terminal 1: Run service
go run main.go

# Terminal 2: Forward webhooks
stripe listen --forward-to localhost:8088/webhook/stripe

# Terminal 3: Trigger test event
stripe trigger payment_intent.succeeded
```

### Test Cards
- Success: `4242 4242 4242 4242`
- Decline: `4000 0000 0000 0002`
- Requires Authentication: `4000 0025 0000 3155`

## Usage

- The payment service listens for incoming payment requests via HTTP and processes them using the integrated payment gateway.
- It communicates with the order service through Kafka, sending payment events and receiving payment-related messages.

## Contributing

Contributions are welcome! Please submit a pull request or open an issue for any enhancements or bug fixes.