package service

import (
	"context"
	"time"

	logger "order-service/log"

	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// InitHealthCheck initializes the health check service
func InitHealthCheck() *health.Server {
	healthServer := health.NewServer()

	// Mark the order service as serving
	healthServer.SetServingStatus("order.OrderService", healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	logger.Logger.Info("Health check service initialized")

	return healthServer
}

// MonitorServiceHealth periodically checks the health of the service
func MonitorServiceHealth(healthServer *health.Server) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Check if gRPC clients are healthy
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		grpcClients := GetGRPCClients()

		cartHealthy := grpcClients.CheckCartServiceHealth(ctx) == nil
		productHealthy := grpcClients.CheckProductServiceHealth(ctx) == nil

		cancel()

		if cartHealthy && productHealthy {
			healthServer.SetServingStatus("order.OrderService", healthpb.HealthCheckResponse_SERVING)
			logger.Logger.Debug("Service health check passed")
		} else {
			healthServer.SetServingStatus("order.OrderService", healthpb.HealthCheckResponse_NOT_SERVING)
			logger.Logger.Warn("Service health check failed",
				"cart_healthy", cartHealthy,
				"product_healthy", productHealthy,
			)
		}
	}
}
