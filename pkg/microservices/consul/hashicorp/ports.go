package pkgconsul

import "github.com/hashicorp/consul/api"

// ConsulClient define la interfaz para interactuar con el cliente de Consul
type Client interface {
	Client() *api.Client
	Address() string // Añadir el método Address a la interfaz
}
