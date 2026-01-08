package pkcresty

import (
	"time"

	"github.com/go-resty/resty/v2"
)

// Client expone las operaciones principales del cliente Resty.
type Client interface {
	// GetClient retorna el cliente interno de Resty.
	GetClient() *resty.Client
	// DoGet es un método de conveniencia para realizar peticiones GET.
	DoGet(url string, result any) (*resty.Response, error)
}

// Config expone las operaciones para configurar el cliente Resty.
type Config interface {
	GetBaseURL() string
	SetBaseURL(string)
	GetTimeout() time.Duration
	SetTimeout(time.Duration)
	Validate() error
}
