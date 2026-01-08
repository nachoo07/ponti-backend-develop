package pkggrpcserver

import (
	"context"
	"fmt"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

var (
	instance Server
	once     sync.Once
	initErr  error
)

type server struct {
	server   *grpc.Server
	listener net.Listener
}

func newServer(config Config) (Server, error) {
	once.Do(func() {
		var opts []grpc.ServerOption
		if config.GetTLSConfig() != nil {
			tlsConfig, err := loadTLSConfig(config.GetTLSConfig())
			if err != nil {
				initErr = fmt.Errorf("failed to load TLS config: %v", err)
				return
			}
			creds := credentials.NewTLS(tlsConfig)
			opts = append(opts, grpc.Creds(creds))
		}

		address := fmt.Sprintf("%s:%d", config.GetHost(), config.GetPort())
		lis, err := net.Listen("tcp", address)
		if err != nil {
			initErr = fmt.Errorf("failed to listen: %v", err)
			return
		}

		srv := grpc.NewServer(opts...)
		reflection.Register(srv) // Registro de reflexión gRPC

		instance = &server{server: srv, listener: lis}
	})
	return instance, initErr
}

func (s *server) Start(ctx context.Context) error {
	go func() {
		<-ctx.Done() // Esperar a que se cancele el contexto
		s.Stop()     // Detener el servidor si se cancela
	}()
	return s.server.Serve(s.listener)
}

func (s *server) Stop() error {
	s.server.GracefulStop()
	return s.listener.Close()
}

func (s *server) RegisterService(ctx context.Context, serviceDesc any, impl any) {
	sd, ok := serviceDesc.(*grpc.ServiceDesc)
	if !ok {
		panic("serviceDesc must be of type *grpc.ServiceDesc")
	}

	// Registrar el servicio con el servidor gRPC
	s.server.RegisterService(sd, impl)
}
