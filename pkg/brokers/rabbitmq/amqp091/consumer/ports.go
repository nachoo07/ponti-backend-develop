package pkgrabbit

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
)

// Config representa la configuración necesaria para conectar a RabbitMQ.
type Config interface {
	GetHost() string
	SetHost(string)

	GetPort() int
	SetPort(int)

	GetUser() string
	SetUser(string)

	GetPassword() string
	SetPassword(string)

	GetVHost() string
	SetVHost(string)

	GetExchange() string
	SetExchange(string)

	GetExchangeType() string
	SetExchangeType(string)

	IsDurable() bool
	SetDurable(bool)

	IsAutoDelete() bool
	SetAutoDelete(bool)

	IsInternal() bool
	SetInternal(bool)

	IsNoWait() bool
	SetNoWait(bool)

	GetConfirmBufferSize() int
	SetConfirmBufferSize(int)

	Validate() error
}

// Consumer define la interfaz para un consumidor de RabbitMQ.
type Consumer interface {
	// Consume inicia la recepción de mensajes de la cola indicada y ejecuta el handler para cada mensaje.
	Consume(ctx context.Context, queueName, consumerTag string, handler func(amqp091.Delivery) error) error
	// Close cierra el canal y la conexión del consumidor.
	Close() error
}

// Logger define la interfaz mínima para realizar logging.
type Logger interface {
	Printf(format string, v ...any)
}
