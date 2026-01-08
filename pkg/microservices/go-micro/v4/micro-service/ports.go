package pkggomicro

import (
	"go-micro.dev/v4"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/server"
)

type Service interface {
	Run() error
	GetService() micro.Service
}

type Config interface {
	GetServer() server.Server
	GetClient() client.Client
	GetConsulAddress() string
	Validate() error
}
