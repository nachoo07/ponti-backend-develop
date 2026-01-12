package pkgafka

import (
	"context"
	"fmt"
	"sync"

	"github.com/segmentio/kafka-go"
)

type service struct {
	config Config
	writer *kafka.Writer
	reader *kafka.Reader
}

func newService(c Config) (Service, error) {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  c.GetBrokers(),
		Balancer: &kafka.LeastBytes{},
	})

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: c.GetBrokers(),
		GroupID: c.GetGroupID(),
	})

	return &service{
		config: c,
		writer: writer,
		reader: reader,
	}, nil
}

func (s *service) Publish(ctx context.Context, topic string, key, value []byte) error {
	msg := kafka.Message{
		Topic: topic,
		Key:   key,
		Value: value,
	}
	return s.writer.WriteMessages(ctx, msg)
}

func (s *service) Consume(ctx context.Context, topics []string, handler func(key, value []byte) error) error {
	var wg sync.WaitGroup
	errorsCh := make(chan error, len(topics))

	for _, topic := range topics {
		wg.Add(1)
		go func(topic string) {
			defer wg.Done()
			reader := kafka.NewReader(kafka.ReaderConfig{
				Brokers: s.config.GetBrokers(),
				GroupID: s.config.GetGroupID(),
				Topic:   topic, // Usamos 'Topic' en singular
			})
			defer reader.Close()

			for {
				m, err := reader.FetchMessage(ctx)
				if err != nil {
					errorsCh <- fmt.Errorf("error fetching message from topic %s: %w", topic, err)
					return
				}

				if err := handler(m.Key, m.Value); err != nil {
					errorsCh <- fmt.Errorf("error handling message from topic %s: %w", topic, err)
					return
				}

				if err := reader.CommitMessages(ctx, m); err != nil {
					errorsCh <- fmt.Errorf("error committing message from topic %s: %w", topic, err)
					return
				}
			}
		}(topic)
	}

	go func() {
		wg.Wait()
		close(errorsCh)
	}()

	// Manejar errores y esperar hasta que se completen todos los goroutines
	for err := range errorsCh {
		if err != nil {
			return err
		}
	}

	return nil
}
