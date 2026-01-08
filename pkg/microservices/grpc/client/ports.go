package pkgcgrpcclient

import (
	"context"

	"google.golang.org/grpc"
)

type Config interface {
	GetHost() string
	SetHost(host string)
	GetPort() int
	SetPort(port int)
	GetTLSConfig() *TLSConfig
	SetTLSConfig(tlsConfig *TLSConfig)
	Validate() error
}

type TLSConfig struct {
	CertFile string
	KeyFile  string
	CAFile   string
}

type Client interface {
	InvokeMethod(ctx context.Context, method string, request, response any) error
	Close() error
	GetConnection() (*grpc.ClientConn, error) // Añadir este método
}
