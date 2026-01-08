package pkgcgrpcclient

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	instance Client
	once     sync.Once
	initErr  error
)

// client structure representing a gRPC client
type client struct {
	conn *grpc.ClientConn
}

// newClient creates a new instance of a gRPC client
func newClient(config Config) (Client, error) {
	once.Do(func() {
		var opts []grpc.DialOption
		if config.GetTLSConfig() != nil {
			tlsConfig, err := loadTLSConfig(config.GetTLSConfig())
			if err != nil {
				initErr = fmt.Errorf("failed to load TLS config: %v", err)
				return
			}
			creds := credentials.NewTLS(tlsConfig)
			opts = append(opts, grpc.WithTransportCredentials(creds))
		} else {
			opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		}

		conn, err := grpc.NewClient(fmt.Sprintf("%s:%d", config.GetHost(), config.GetPort()), opts...)
		if err != nil {
			initErr = fmt.Errorf("failed to connect to gRPC server: %v", err)
			return
		}

		instance = &client{conn: conn}
	})
	return instance, initErr
}

// Implementation of GetConnection
func (client *client) GetConnection() (*grpc.ClientConn, error) {
	if client.conn == nil {
		return nil, fmt.Errorf("gRPC client connection is not initialized")
	}
	return client.conn, nil
}

// InvokeMethod invokes a gRPC method
func (client *client) InvokeMethod(ctx context.Context, method string, request, response any) error {
	// Additional check to avoid invoking with a nil connection
	if client.conn == nil {
		return fmt.Errorf("gRPC client connection is not initialized")
	}
	return client.conn.Invoke(ctx, method, request, response)
}

// Close closes the gRPC client connection
func (client *client) Close() error {
	if client.conn == nil {
		return fmt.Errorf("gRPC client connection is not initialized")
	}
	return client.conn.Close()
}
