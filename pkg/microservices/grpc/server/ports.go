package pkggrpcserver

import "context"

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

type Server interface {
	Start(context.Context) error
	Stop() error
	RegisterService(context.Context, any, any)
}
