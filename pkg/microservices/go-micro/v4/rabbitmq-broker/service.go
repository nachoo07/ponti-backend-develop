package pkgrabbit

import (
	"context"
	"fmt"
	"sync"

	"github.com/rabbitmq/amqp091-go"
	"go-micro.dev/v4/broker"
	"go-micro.dev/v4/logger"
)

var (
	instance  Broker
	once      sync.Once
	initError error
)

// service estructura que implementa la interfaz Broker
type service struct {
	b  broker.Broker
	mu sync.Mutex
}

// newBroker crea una nueva instancia del servicio que implementa Broker
func newService(config Config) (Broker, error) {
	once.Do(func() {
		// Configurar el broker para usar RabbitMQ
		rabbit := broker.NewBroker(
			broker.Addrs(config.GetAddress()),
		)

		// Inicializar el broker
		if err := rabbit.Init(); err != nil {
			initError = fmt.Errorf("error al inicializar el broker: %w", err)
			return
		}

		// Conectar al broker
		if err := rabbit.Connect(); err != nil {
			initError = fmt.Errorf("error al conectar con el broker: %w", err)
			return
		}

		instance = &service{
			b: rabbit,
		}
	})

	if initError != nil {
		return nil, initError
	}

	return instance, nil
}

func (s *service) Publish(targetType, targetName, routingKey string, body []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg := &broker.Message{
		Header: map[string]string{
			"Content-Type": "application/json",
			"routingKey":   routingKey,
		},
		Body: body,
	}

	if err := s.b.Publish(targetName, msg); err != nil {
		return fmt.Errorf("error al publicar mensaje: %w", err)
	}

	logger.Infof("Mensaje publicado en %s con routing key %s", targetName, routingKey)
	return nil
}

// Subscribe se suscribe a un tema y procesa los mensajes entrantes
func (s *service) Subscribe(ctx context.Context, handler func(amqp091.Delivery), topic string) error {
	// Llamar a Subscribe y capturar ambos valores
	_, err := s.b.Subscribe(topic, func(event broker.Event) error {
		msg := event.Message()

		// Convertir broker.Message a amqp091.Delivery
		delivery := amqp091.Delivery{
			Body: msg.Body,
		}

		// Llamar al handler con el mensaje recibido
		handler(delivery)
		return nil
	}, broker.Queue("consumer-queue"))

	if err != nil {
		return fmt.Errorf("error al suscribirse: %w", err)
	}

	return nil
}

// Close cierra la conexión con el broker
func (s *service) Close() error {
	return s.b.Disconnect()
}

// GetBroker devuelve la instancia del broker
func (s *service) GetBroker() broker.Broker {
	return s.b
}
