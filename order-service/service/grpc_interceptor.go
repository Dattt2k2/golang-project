package service

import (
	"context"
	"sync"
	"time"

	logger "order-service/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Semaphore for controlling concurrent requests
type Semaphore struct {
	sem chan struct{}
	mu  sync.Mutex
}

func NewSemaphore(maxConcurrent int) *Semaphore {
	return &Semaphore{
		sem: make(chan struct{}, maxConcurrent),
	}
}

func (s *Semaphore) Acquire(ctx context.Context) error {
	select {
	case s.sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return status.Error(codes.ResourceExhausted, "too many concurrent requests")
	}
}

func (s *Semaphore) Release() {
	<-s.sem
}

// Global semaphore for rate limiting
var globalSemaphore = NewSemaphore(500) // Allow max 500 concurrent requests

// UnaryServerInterceptor creates a unary interceptor for rate limiting and logging
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Try to acquire semaphore
		if err := globalSemaphore.Acquire(ctx); err != nil {
			logger.Logger.Warn("Request rejected due to rate limiting",
				"method", info.FullMethod,
			)
			return nil, err
		}
		defer globalSemaphore.Release()

		// Add timeout to context if not already set
		if _, ok := ctx.Deadline(); !ok {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
			defer cancel()
		}

		// Call the handler
		resp, err := handler(ctx, req)

		// Log the request
		duration := time.Since(start)
		if err != nil {
			logger.Logger.Error("gRPC request failed",
				"method", info.FullMethod,
				"duration_ms", duration.Milliseconds(),
				"error", err.Error(),
			)
		} else {
			logger.Logger.Info("gRPC request completed",
				"method", info.FullMethod,
				"duration_ms", duration.Milliseconds(),
			)
		}

		return resp, err
	}
}

// StreamServerInterceptor creates a stream interceptor for rate limiting
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := ss.Context()

		// Try to acquire semaphore
		if err := globalSemaphore.Acquire(ctx); err != nil {
			logger.Logger.Warn("Stream request rejected due to rate limiting",
				"method", info.FullMethod,
			)
			return err
		}
		defer globalSemaphore.Release()

		start := time.Now()
		err := handler(srv, ss)
		duration := time.Since(start)

		if err != nil {
			logger.Logger.Error("gRPC stream failed",
				"method", info.FullMethod,
				"duration_ms", duration.Milliseconds(),
				"error", err.Error(),
			)
		} else {
			logger.Logger.Info("gRPC stream completed",
				"method", info.FullMethod,
				"duration_ms", duration.Milliseconds(),
			)
		}

		return err
	}
}
