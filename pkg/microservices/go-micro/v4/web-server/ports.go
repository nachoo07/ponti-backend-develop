package pkggomicro

type Server interface {
	Run() error
	SetRouter(router any) error
}

type Config interface {
	GetServerName() string
	GetServerHost() string
	GetServerPort() int
	GetServerID() string
	GetServerAddress() string
	GetConsulAddress() string
	GetRouter() any
	Validate() error
}
