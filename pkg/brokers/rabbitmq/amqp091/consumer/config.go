// config.go
package pkgrabbit

import (
	"fmt"
)

type config struct {
	host              string
	port              int
	user              string
	password          string
	vhost             string
	exchange          string
	exchangeType      string
	durable           bool
	autoDelete        bool
	internal          bool
	noWait            bool
	confirmBufferSize int
}

// newConfig crea una nueva instancia de Config. Si confirmBufferSize es menor o igual a 0 se asigna un valor por defecto.
func newConfig(host string, port int, user, password, vhost, exchange, exchangeType string,
	durable, autoDelete, internal, noWait bool, confirmBufferSize int) Config {
	if confirmBufferSize <= 0 {
		confirmBufferSize = 10
	}
	return &config{
		host:              host,
		port:              port,
		user:              user,
		password:          password,
		vhost:             vhost,
		exchange:          exchange,
		exchangeType:      exchangeType,
		durable:           durable,
		autoDelete:        autoDelete,
		internal:          internal,
		noWait:            noWait,
		confirmBufferSize: confirmBufferSize,
	}
}

func (c *config) GetHost() string     { return c.host }
func (c *config) SetHost(host string) { c.host = host }

func (c *config) GetPort() int     { return c.port }
func (c *config) SetPort(port int) { c.port = port }

func (c *config) GetUser() string     { return c.user }
func (c *config) SetUser(user string) { c.user = user }

func (c *config) GetPassword() string         { return c.password }
func (c *config) SetPassword(password string) { c.password = password }

func (c *config) GetVHost() string      { return c.vhost }
func (c *config) SetVHost(vhost string) { c.vhost = vhost }

func (c *config) GetExchange() string         { return c.exchange }
func (c *config) SetExchange(exchange string) { c.exchange = exchange }

func (c *config) GetExchangeType() string             { return c.exchangeType }
func (c *config) SetExchangeType(exchangeType string) { c.exchangeType = exchangeType }

func (c *config) IsDurable() bool         { return c.durable }
func (c *config) SetDurable(durable bool) { c.durable = durable }

func (c *config) IsAutoDelete() bool            { return c.autoDelete }
func (c *config) SetAutoDelete(autoDelete bool) { c.autoDelete = autoDelete }

func (c *config) IsInternal() bool          { return c.internal }
func (c *config) SetInternal(internal bool) { c.internal = internal }

func (c *config) IsNoWait() bool        { return c.noWait }
func (c *config) SetNoWait(noWait bool) { c.noWait = noWait }

func (c *config) GetConfirmBufferSize() int { return c.confirmBufferSize }
func (c *config) SetConfirmBufferSize(size int) {
	if size > 0 {
		c.confirmBufferSize = size
	}
}

func (c *config) Validate() error {
	if c.host == "" {
		return fmt.Errorf("rabbitmq host is not configured")
	}
	if c.port == 0 {
		return fmt.Errorf("rabbitmq port is not configured")
	}
	if c.user == "" {
		return fmt.Errorf("rabbitmq user is not configured")
	}
	if c.password == "" {
		return fmt.Errorf("rabbitmq password is not configured")
	}
	if c.vhost == "" {
		return fmt.Errorf("rabbitmq vhost is not configured")
	}
	if c.exchange == "" {
		return fmt.Errorf("rabbitmq exchange is not configured")
	}
	if c.exchangeType == "" {
		return fmt.Errorf("rabbitmq exchange type is not configured")
	}
	return nil
}
