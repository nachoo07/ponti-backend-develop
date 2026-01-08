package pkggomicro

import (
	"fmt"
	"sync"

	"github.com/go-micro/plugins/v4/server/grpc"
	gmserver "go-micro.dev/v4/server"
)

var (
	instance  Server
	once      sync.Once
	initError error
)

type grpcserver struct {
	s gmserver.Server
}

func newServer(config Config) (Server, error) {
	once.Do(func() {
		srv, err := setupServer(config)
		if err != nil {
			initError = fmt.Errorf("error setting up grpcserver: %w", err)
			return
		}
		instance = &grpcserver{
			s: srv,
		}
	})

	if initError != nil {
		return nil, initError
	}

	return instance, nil
}

func setupServer(config Config) (gmserver.Server, error) {
	s := grpc.NewServer(
		gmserver.Name(config.GetServerName()),
		gmserver.Id(config.GetServerID()),
		gmserver.Address(fmt.Sprintf("%s:%d", config.GetServerHost(), config.GetServerPort())),
	)

	return s, nil
}

func (s *grpcserver) GetServer() gmserver.Server {
	return s.s
}
