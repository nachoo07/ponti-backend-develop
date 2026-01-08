package pkggomicro

import "go-micro.dev/v4/server"

type Server interface {
	GetServer() server.Server
}

type Config interface {
	GetServerName() string
	GetServerHost() string
	GetServerPort() int
	GetServerID() string
	Validate() error
}
