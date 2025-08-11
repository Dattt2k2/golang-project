# Payment Service

This project is a payment service that integrates with a payment gateway and uses Kafka for communication with the order service. It is designed to handle payment processing and manage payment-related events efficiently.

## Project Structure

```
payment-service
├── src
│   ├── main.go                # Entry point of the application
│   ├── service                # Contains service-related logic
│   │   ├── paymentGateway.go   # Payment gateway integration
│   │   ├── kafkaProducer.go    # Kafka producer implementation
│   │   ├── kafkaConsumer.go    # Kafka consumer implementation
│   │   └── grpcConnection.go    # gRPC connection to order service
│   ├── config                 # Configuration settings
│   │   └── config.go          # Configuration struct and settings
│   ├── handlers               # HTTP request handlers
│   │   └── paymentHandler.go   # Payment request handling
│   ├── models                 # Data structures
│   │   └── payment.go         # Payment-related structs
│   └── utils                  # Utility functions
│       └── logger.go          # Logging utilities
├── go.mod                     # Module definition and dependencies
├── go.sum                     # Dependency checksums
└── README.md                  # Project documentation
```

## Setup Instructions

1. **Clone the repository:**
   ```
   git clone <repository-url>
   cd payment-service
   ```

2. **Install dependencies:**
   ```
   go mod tidy
   ```

3. **Configure the service:**
   Update the `src/config/config.go` file with your Kafka broker addresses and payment gateway credentials.

4. **Run the service:**
   ```
   go run src/main.go
   ```

## Usage

- The payment service listens for incoming payment requests via HTTP and processes them using the integrated payment gateway.
- It communicates with the order service through Kafka, sending payment events and receiving payment-related messages.

## Contributing

Contributions are welcome! Please submit a pull request or open an issue for any enhancements or bug fixes.