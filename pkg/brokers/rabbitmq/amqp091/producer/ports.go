package pkgrabbit

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
)

// Producer define la interfaz para el productor de RabbitMQ.
type Producer interface {
	// Channel devuelve el canal actual de RabbitMQ.
	Channel() (*amqp091.Channel, error)
	// Close cierra de manera segura el productor de RabbitMQ.
	Close() error
	// Produce envía un mensaje a la cola especificada con opción de reply-to e ID de correlación.
	Produce(ctx context.Context, queueName, replyTo, corrID string, message any) (string, error)
	// ProduceWithRetry envía un mensaje con reintentos en caso de fallo.
	ProduceWithRetry(ctx context.Context, queueName, replyTo, corrID string, message any, maxRetries int) (string, error)
	// GetConnection devuelve la conexión actual a RabbitMQ.
	GetConnection() *amqp091.Connection
}

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

	// Tamaño del buffer para confirmaciones, opcional.
	GetConfirmBufferSize() int
	SetConfirmBufferSize(int)

	Validate() error
}

// Logger es una interfaz mínima para logging. Se puede inyectar un logger más sofisticado.
type Logger interface {
	Printf(format string, v ...any)
}
