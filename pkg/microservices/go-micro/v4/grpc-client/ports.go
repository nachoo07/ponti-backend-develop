package pkggomicro

import "go-micro.dev/v4/client"

type Client interface {
	GetClient() client.Client
	GetServerName() string
}

type Config interface {
	Validate() error
	GetConsulAddress() string
	GetServerName() string
}
