package pkgrabbit

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
)

// Logger define la interfaz mínima para realizar logging.
type Logger interface {
	Printf(format string, v ...any)
}

// Service define las operaciones de un broker RabbitMQ que actúa como productor y consumidor.
type Service interface {
	// Publish envía un mensaje al destino indicado (exchange o queue).
	Publish(targetType, targetName, routingKey string, body []byte) error
	// Subscribe se suscribe al destino indicado y retorna un canal de entregas.
	Subscribe(ctx context.Context, targetType, targetName, exchangeType, routingKey string) (<-chan amqp091.Delivery, error)
	// SetupExchangeAndQueue configura el exchange y la cola y enlaza la cola con el exchange usando la routing key.
	SetupExchangeAndQueue(exchangeName, exchangeType, queueName, routingKey string) error
	// Close cierra de forma segura el canal y la conexión.
	Close() error
	// GetConnection retorna la conexión actual a RabbitMQ.
	GetConnection() *amqp091.Connection
}

// Config define la configuración específica para RabbitMQ que utilizará el broker.
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

	GetQueue() string
	SetQueue(string)

	GetExchange() string
	SetExchange(string)

	GetExchangeType() string
	SetExchangeType(string)

	GetRoutingKey() string
	SetRoutingKey(string)

	GetAutoAck() bool
	SetAutoAck(bool)

	GetExclusive() bool
	SetExclusive(bool)

	GetNoLocal() bool
	SetNoLocal(bool)

	GetNoWait() bool
	SetNoWait(bool)

	// GetAddress genera la URL de conexión para RabbitMQ.
	GetAddress() string

	Validate() error
}
