package service

import (
    "context"
    "sync"
    "time"

    cartPb "module/gRPC-cart/service"
    productPb "module/gRPC-Product/service"
    logger "order-service/log"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/grpc/keepalive"
	"fmt"
)

type GRPCClients struct {
    CartClient    cartPb.CartServiceClient
    ProductClient productPb.ProductServiceClient
    cartConn      *grpc.ClientConn
    productConn   *grpc.ClientConn
    mu            sync.RWMutex
}

var (
    grpcClients *GRPCClients
    once        sync.Once
)

// GetGRPCClients returns singleton instance of gRPC clients
func GetGRPCClients() *GRPCClients {
    once.Do(func() {
        grpcClients = &GRPCClients{}
        grpcClients.initConnections()
    })
    return grpcClients
}

func (g *GRPCClients) initConnections() {
    g.mu.Lock()
    defer g.mu.Unlock()

    // Cart service connection
    cartConn, err := g.createConnection("cart-service:8090", "Cart-service")
    if err != nil {
        logger.Err("Failed to connect to Cart-service", err)
        return
    }
    g.cartConn = cartConn
    g.CartClient = cartPb.NewCartServiceClient(cartConn)

    // Product service connection
    productConn, err := g.createConnection("product-service:8089", "Product-service")
    if err != nil {
        logger.Err("Failed to connect to Product-service", err)
        return
    }
    g.productConn = productConn
    g.ProductClient = productPb.NewProductServiceClient(productConn)

}

func (g *GRPCClients) createConnection(address, serviceName string) (*grpc.ClientConn, error) {

    keepAliveParams := keepalive.ClientParameters{
        Time:                10 * time.Second,
        Timeout:             3 * time.Second,
        PermitWithoutStream: true,
    }

    conn, err := grpc.NewClient(address,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithKeepaliveParams(keepAliveParams),
    )

    if err != nil {
        logger.Err("Failed to connect to "+serviceName, err)
        return nil, err
    }

    return conn, nil
}

// GetCartClient returns cart service client with connection check
func (g *GRPCClients) GetCartClient() cartPb.CartServiceClient {
    g.mu.RLock()
    defer g.mu.RUnlock()
    
    if g.CartClient == nil {
        logger.Logger.Warn("Cart client is not initialized")
        return nil
    }
    return g.CartClient
}

// GetProductClient returns product service client with connection check
func (g *GRPCClients) GetProductClient() productPb.ProductServiceClient {
    g.mu.RLock()
    defer g.mu.RUnlock()
    
    if g.ProductClient == nil {
        logger.Logger.Warn("Product client is not initialized")
        return nil 
    }
    return g.ProductClient
}

// CloseConnections closes all gRPC connections
func (g *GRPCClients) CloseConnections() {
    g.mu.Lock()
    defer g.mu.Unlock()

    if g.cartConn != nil {
        if err := g.cartConn.Close(); err != nil {
            logger.Err("Error closing cart service connection", err)
        }
    }

    if g.productConn != nil {
        if err := g.productConn.Close(); err != nil {
            logger.Err("Error closing product service connection", err)
        }
    }
}

// Legacy methods for backward compatibility
func CartServiceConnection() cartPb.CartServiceClient {
    clients := GetGRPCClients()
    return clients.GetCartClient()
}

func ProductServiceConnection() productPb.ProductServiceClient {
    clients := GetGRPCClients()
    return clients.GetProductClient()
}

// Health check methods
func (g *GRPCClients) CheckCartServiceHealth(ctx context.Context) error {
    client := g.GetCartClient()
    if client == nil {
        return fmt.Errorf("cart client not available")
    }
    
    // You can implement a health check RPC call here
    // For now, just check if client exists
    return nil
}

func (g *GRPCClients) CheckProductServiceHealth(ctx context.Context) error {
    client := g.GetProductClient()
    if client == nil {
        return fmt.Errorf("product client not available")
    }
    
    // You can implement a health check RPC call here
    return nil
}