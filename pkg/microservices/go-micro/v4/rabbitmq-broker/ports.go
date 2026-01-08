package pkgrabbit

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
	"go-micro.dev/v4/broker"
)

// Broker define las operaciones para un sistema de mensajería RabbitMQ.
type Broker interface {
	Publish(string, string, string, []byte) error
	Subscribe(context.Context, func(amqp091.Delivery), string) error
	Close() error
	GetBroker() broker.Broker
}

// Config define la configuración específica para RabbitMQ.
type Config interface {
	GetName() string
	SetName(string)

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

	GetAddress() string

	Validate() error
}
