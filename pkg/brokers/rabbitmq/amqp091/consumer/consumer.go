// consumer.go
package pkgrabbit

import (
	"context"
	"fmt"
	"sync"

	"github.com/rabbitmq/amqp091-go"
)



type consumer struct {
	conn     *amqp091.Connection
	channel  *amqp091.Channel
	config   Config
	logger   Logger
	reconnMu sync.Mutex // Para proteger la reconexión
}

// NewConsumer crea una nueva instancia de Consumer utilizando la configuración y un logger.
// Si logger es nil se usa el logger por defecto.
func NewConsumer(config Config, logger Logger) (Consumer, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}
	if logger == nil {
		logger = &defaultLogger{}
	}
	connString := fmt.Sprintf("amqp://%s:%s@%s:%d%s",
		config.GetUser(), config.GetPassword(), config.GetHost(), config.GetPort(), config.GetVHost())
	conn, err := amqp091.Dial(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}
	return &consumer{
		conn:    conn,
		channel: ch,
		config:  config,
		logger:  logger,
	}, nil
}

// reconnect intenta restablecer la conexión y el canal de forma segura.
func (c *consumer) reconnect() error {
	c.reconnMu.Lock()
	defer c.reconnMu.Unlock()

	c.logger.Printf("Attempting consumer reconnection...")

	// Cerrar recursos actuales (si existen).
	if c.channel != nil {
		_ = c.channel.Close()
	}
	if c.conn != nil {
		_ = c.conn.Close()
	}

	connString := fmt.Sprintf("amqp://%s:%s@%s:%d%s",
		c.config.GetUser(), c.config.GetPassword(), c.config.GetHost(), c.config.GetPort(), c.config.GetVHost())
	conn, err := amqp091.Dial(connString)
	if err != nil {
		return fmt.Errorf("failed to reconnect to RabbitMQ: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel after reconnection: %w", err)
	}
	c.conn = conn
	c.channel = ch

	c.logger.Printf("Consumer reconnection successful")
	return nil
}

// Consume inicia el consumo de mensajes de la cola indicada.
// Si el canal se cierra, intenta reconectarse y re-registrar el consumidor.
func (c *consumer) Consume(ctx context.Context, queueName, consumerTag string, handler func(amqp091.Delivery) error) error {
	msgs, err := c.channel.Consume(
		queueName,
		consumerTag,
		false, // Acknowledge manual
		false, // No exclusive
		false, // No local
		false, // No wait
		nil,   // Argumentos
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case d, ok := <-msgs:
			if !ok {
				// El canal se cerró; intentar reconectar y re-registrar el consumidor.
				c.logger.Printf("Consumer channel closed, attempting reconnection...")
				if recErr := c.reconnect(); recErr != nil {
					return fmt.Errorf("failed to reconnect consumer: %w", recErr)
				}
				newMsgs, err := c.channel.Consume(queueName, consumerTag, false, false, false, false, nil)
				if err != nil {
					return fmt.Errorf("failed to re-register consumer after reconnection: %w", err)
				}
				msgs = newMsgs
				continue
			}
			// Procesar el mensaje usando el handler proporcionado.
			if err := handler(d); err != nil {
				c.logger.Printf("Error processing message: %v", err)
				// Rechazar el mensaje y pedir reenvío.
				d.Nack(false, true)
			} else {
				d.Ack(false)
			}
		}
	}
}

// Close cierra el canal y la conexión del consumidor.
func (c *consumer) Close() error {
	var errs []error
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close consumer channel: %w", err))
		}
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close consumer connection: %w", err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors closing consumer: %v", errs)
	}
	return nil
}
