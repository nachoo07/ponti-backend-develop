// producer.go
package pkgrabbit

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

// producer implementa la interfaz Producer para RabbitMQ.
type producer struct {
	conn      *amqp091.Connection
	channel   *amqp091.Channel
	exchange  string
	confirmCh chan amqp091.Confirmation // Canal para confirmaciones.
	config    Config
	reconnMu  sync.Mutex // Protege la reconexión.
	publishMu sync.Mutex // Protege la publicación concurrente.
	logger    Logger
}

// newProducer crea una nueva instancia de Producer. Se inyecta la configuración y, opcionalmente, un logger.
// Si logger es nil se utiliza el logger por defecto.
func newProducer(config Config, logger Logger) (Producer, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}
	if logger == nil {
		logger = log.Default()
	}

	p, err := createProducer(config, logger)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// createProducer encapsula la lógica de creación, declarando el exchange y configurando el canal.
func createProducer(config Config, logger Logger) (*producer, error) {
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

	// Declarar el exchange según la configuración.
	err = ch.ExchangeDeclare(
		config.GetExchange(),     // Nombre del exchange.
		config.GetExchangeType(), // Tipo de exchange.
		config.IsDurable(),       // Durable.
		config.IsAutoDelete(),    // Auto-delete.
		config.IsInternal(),      // Internal.
		config.IsNoWait(),        // No-wait.
		nil,                      // Argumentos adicionales.
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Habilitar el modo de confirmación del publicador.
	err = ch.Confirm(false)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to put channel into confirm mode: %w", err)
	}

	confirmCh := ch.NotifyPublish(make(chan amqp091.Confirmation, config.GetConfirmBufferSize()))

	return &producer{
		conn:      conn,
		channel:   ch,
		exchange:  config.GetExchange(),
		confirmCh: confirmCh,
		config:    config,
		logger:    logger,
	}, nil
}

// reconnect intenta restablecer la conexión y el canal de forma segura.
func (p *producer) reconnect() error {
	p.reconnMu.Lock()
	defer p.reconnMu.Unlock()

	p.logger.Printf("Attempting reconnection...")

	// Cerrar el canal y la conexión actuales.
	if err := p.channel.Close(); err != nil {
		p.logger.Printf("failed to close channel during reconnect: %v", err)
	}
	if err := p.conn.Close(); err != nil {
		p.logger.Printf("failed to close connection during reconnect: %v", err)
	}

	// Intentar recrear el productor.
	newProd, err := createProducer(p.config, p.logger)
	if err != nil {
		return fmt.Errorf("reconnect failed: %w", err)
	}

	// Actualizar los recursos internos de forma atómica.
	p.publishMu.Lock()
	defer p.publishMu.Unlock()
	p.conn = newProd.conn
	p.channel = newProd.channel
	p.exchange = newProd.exchange
	p.confirmCh = newProd.confirmCh

	p.logger.Printf("Reconnection successful")
	return nil
}

// Produce envía un mensaje a la cola especificada. Se protege el acceso concurrente al canal.
func (p *producer) Produce(ctx context.Context, queueName, replyTo, corrID string, message any) (string, error) {
	// Convertir el mensaje a JSON.
	body, err := json.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("failed to marshal message: %w", err)
	}

	// Generar un ID de correlación si no se proporciona.
	if corrID == "" {
		corrID = fmt.Sprintf("%d", time.Now().UnixNano())
	}

	// Proteger la publicación.
	p.publishMu.Lock()
	err = p.channel.PublishWithContext(ctx,
		p.exchange, // Exchange.
		queueName,  // Routing key (cola o binding key).
		false,      // Mandatory.
		false,      // Immediate.
		amqp091.Publishing{
			ContentType:   "application/json",
			Body:          body,
			CorrelationId: corrID,
			ReplyTo:       replyTo,
		})
	p.publishMu.Unlock()

	if err != nil {
		p.logger.Printf("publish error: %v; attempting reconnect", err)
		// Intentar reconectar.
		if recErr := p.reconnect(); recErr != nil {
			return "", fmt.Errorf("failed to publish message and reconnect: %w", err)
		}
		// Reintentar la publicación una vez reconectado.
		p.publishMu.Lock()
		err = p.channel.PublishWithContext(ctx,
			p.exchange, queueName, false, false,
			amqp091.Publishing{
				ContentType:   "application/json",
				Body:          body,
				CorrelationId: corrID,
				ReplyTo:       replyTo,
			})
		p.publishMu.Unlock()

		if err != nil {
			return "", fmt.Errorf("failed to publish message after reconnect: %w", err)
		}
	}

	// Esperar la confirmación del mensaje.
	select {
	case confirmation, ok := <-p.confirmCh:
		if !ok {
			return "", fmt.Errorf("confirmation channel closed")
		}
		if confirmation.Ack {
			p.logger.Printf("Message acknowledged by RabbitMQ")
		} else {
			return "", fmt.Errorf("message not acknowledged by RabbitMQ")
		}
	case <-ctx.Done():
		return "", ctx.Err()
	}

	return corrID, nil
}

// ProduceWithRetry envía un mensaje con reintentos y backoff exponencial en caso de fallo.
func (p *producer) ProduceWithRetry(ctx context.Context, queueName, replyTo, corrID string, message any, maxRetries int) (string, error) {
	var err error
	var id string
	for attempt := 0; attempt < maxRetries; attempt++ {
		id, err = p.Produce(ctx, queueName, replyTo, corrID, message)
		if err == nil {
			return id, nil
		}
		p.logger.Printf("Retry %d/%d failed: %v", attempt+1, maxRetries, err)
		sleepDuration := time.Duration(1<<attempt) * time.Second
		select {
		case <-time.After(sleepDuration):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	return "", fmt.Errorf("max retries reached: %w", err)
}

// Channel devuelve el canal actual de RabbitMQ.
func (p *producer) Channel() (*amqp091.Channel, error) {
	if p.channel == nil {
		return nil, fmt.Errorf("RabbitMQ channel is not initialized")
	}
	return p.channel, nil
}

// Close cierra de manera segura el canal y la conexión.
func (p *producer) Close() error {
	var errs []error

	if err := p.channel.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close RabbitMQ channel: %w", err))
	}

	if err := p.conn.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close RabbitMQ connection: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors while closing pkgrabbit: %v", errs)
	}

	return nil
}

// GetConnection devuelve la conexión actual a RabbitMQ.
func (p *producer) GetConnection() *amqp091.Connection {
	return p.conn
}
