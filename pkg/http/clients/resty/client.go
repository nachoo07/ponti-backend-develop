package pkcresty

import (
	"time"

	"github.com/go-resty/resty/v2"
)

// client implementa la interfaz Client.
type client struct {
	restyClient *resty.Client
	config      Config
}

// newClient crea una nueva instancia del cliente Resty sin usar un singleton.
func newClient(config Config) (Client, error) {
	// Creamos el cliente Resty
	r := resty.New()
	r.SetBaseURL(config.GetBaseURL())
	r.SetTimeout(time.Duration(config.GetTimeout()) * time.Second)

	// Retornamos la nueva instancia del cliente
	return &client{
		restyClient: r,
		config:      config,
	}, nil
}

// GetClient retorna el cliente Resty configurado.
func (c *client) GetClient() *resty.Client {
	return c.restyClient
}

// DoGet es un método de conveniencia para realizar peticiones GET.
func (c *client) DoGet(url string, result any) (*resty.Response, error) {
	return c.restyClient.R().SetResult(result).Get(url)
}
