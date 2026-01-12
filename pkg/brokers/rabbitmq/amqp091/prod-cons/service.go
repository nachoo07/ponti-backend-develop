package pkgrabbit

import (
	"context"
	"fmt"
	"sync"

	"github.com/rabbitmq/amqp091-go"
)

type service struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	config  Config
	mutex   sync.Mutex
}

func newService(config Config) (Service, error) {
	conn, err := amqp091.Dial(config.GetAddress())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}
	// Declarar el exchange.
	if err := ch.ExchangeDeclare(
		config.GetExchange(),
		config.GetExchangeType(),
		true,               // Durable
		false,              // Auto-deleted
		false,              // Internal
		config.GetNoWait(), // No-wait
		nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}
	// Declarar la cola.
	if _, err := ch.QueueDeclare(
		config.GetQueue(),
		true,  // Durable
		false, // Delete when unused
		config.GetExclusive(),
		config.GetNoWait(),
		nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}
	// Bind de la cola al exchange.
	if err := ch.QueueBind(
		config.GetQueue(),
		config.GetRoutingKey(),
		config.GetExchange(),
		false,
		nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}
	// Habilitar confirm mode para el publicador.
	if err := ch.Confirm(false); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to enable confirm mode: %w", err)
	}
	return &service{
		conn:    conn,
		channel: ch,
		config:  config,
	}, nil
}

func (s *service) Publish(targetType, targetName, routingKey string, body []byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	publishing := amqp091.Publishing{
		ContentType: "text/plain",
		Body:        body,
	}

	switch targetType {
	case "exchange":
		if err := s.channel.Publish(
			targetName, // Exchange
			routingKey, // Routing key
			false,
			false,
			publishing,
		); err != nil {
			return fmt.Errorf("failed to publish message to exchange: %w", err)
		}
	case "queue":
		if err := s.channel.Publish(
			"",         // Publicar directamente a la cola
			targetName, // Nombre de la cola
			false,
			false,
			publishing,
		); err != nil {
			return fmt.Errorf("failed to publish message to queue: %w", err)
		}
	default:
		return fmt.Errorf("invalid target type: %s", targetType)
	}
	return nil
}

func (s *service) Subscribe(ctx context.Context, targetType, targetName, exchangeType, routingKey string) (<-chan amqp091.Delivery, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if targetType == "exchange" {
		if err := s.channel.ExchangeDeclare(
			targetName,
			exchangeType,
			true,
			false,
			false,
			s.config.GetNoWait(),
			nil,
		); err != nil {
			return nil, fmt.Errorf("failed to declare exchange: %w", err)
		}
	}

	queue, err := s.channel.QueueDeclare(
		s.config.GetQueue(),
		true,
		false,
		s.config.GetExclusive(),
		s.config.GetNoWait(),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	if targetType == "exchange" {
		if err := s.channel.QueueBind(
			queue.Name,
			routingKey,
			targetName,
			false,
			nil,
		); err != nil {
			return nil, fmt.Errorf("failed to bind queue: %w", err)
		}
	}

	msgs, err := s.channel.Consume(
		queue.Name,
		"",
		s.config.GetAutoAck(),
		s.config.GetExclusive(),
		s.config.GetNoLocal(),
		s.config.GetNoWait(),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to consume messages: %w", err)
	}

	filteredMsgs := make(chan amqp091.Delivery)
	go func() {
		defer close(filteredMsgs)
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				filteredMsgs <- msg
			}
		}
	}()

	return filteredMsgs, nil
}

func (s *service) SetupExchangeAndQueue(exchangeName, exchangeType, queueName, routingKey string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if err := s.channel.ExchangeDeclare(
		exchangeName,
		exchangeType,
		true,
		false,
		false,
		s.config.GetNoWait(),
		nil,
	); err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	if _, err := s.channel.QueueDeclare(
		queueName,
		true,
		false,
		s.config.GetExclusive(),
		s.config.GetNoWait(),
		nil,
	); err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	if err := s.channel.QueueBind(
		queueName,
		routingKey,
		exchangeName,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}
	return nil
}

func (s *service) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var errs []error
	if err := s.channel.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close channel: %w", err))
	}
	if err := s.conn.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close connection: %w", err))
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors while closing service: %v", errs)
	}
	return nil
}

func (s *service) GetConnection() *amqp091.Connection {
	return s.conn
}
