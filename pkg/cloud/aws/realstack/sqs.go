package pkgrealstack

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"

	"github.com/alphacodinggroup/ponti-backend/pkg/cloud/aws/ports"
)

// sqsClient implementa la interfaz SQSClient
type sqsClient struct {
	client *sqs.Client
}

// NewSQSClient crea una nueva instancia del cliente SQS
func NewSQSClient(cfg aws.Config) ports.SQSClient {
	return &sqsClient{
		client: sqs.NewFromConfig(cfg),
	}
}

// GetOrCreateQueueURL obtiene o crea una cola SQS
func (c *sqsClient) GetOrCreateQueueURL(ctx context.Context, queueName string) (string, error) {
	if queueName == "" {
		return "", fmt.Errorf("queue name cannot be empty")
	}

	// Primero intentar obtener la URL de la cola existente
	getCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	out, err := c.client.GetQueueUrl(getCtx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})
	if err == nil {
		return *out.QueueUrl, nil
	}

	// Si la cola no existe, crearla con configuración por defecto
	createCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	createOut, err := c.client.CreateQueue(createCtx, &sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
		Attributes: map[string]string{
			"VisibilityTimeout":             fmt.Sprintf("%d", defaultVisibilityTimeout),
			"MessageRetentionPeriod":        "345600", // 4 días
			"ReceiveMessageWaitTimeSeconds": fmt.Sprintf("%d", defaultWaitTimeSeconds),
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create queue %s: %w", queueName, err)
	}

	return *createOut.QueueUrl, nil
}

// SendMessage envía un mensaje a la cola SQS
func (c *sqsClient) SendMessage(ctx context.Context, queueURL string, messageBody string) error {
	if queueURL == "" || messageBody == "" {
		return fmt.Errorf("queue URL and message body cannot be empty")
	}

	sendCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(messageBody),
	}

	_, err := c.client.SendMessage(sendCtx, input)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// ReceiveMessages recibe mensajes de la cola SQS
func (c *sqsClient) ReceiveMessages(ctx context.Context, queueURL string, maxMessages int32) ([]ports.SQSMessage, error) {
	if queueURL == "" {
		return nil, fmt.Errorf("queue URL cannot be empty")
	}

	if maxMessages <= 0 || maxMessages > defaultMaxNumberOfMessages {
		maxMessages = defaultMaxNumberOfMessages
	}

	receiveCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	input := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueURL),
		MaxNumberOfMessages: maxMessages,
		WaitTimeSeconds:     defaultWaitTimeSeconds,
		AttributeNames: []types.QueueAttributeName{
			types.QueueAttributeNameAll,
		},
		MessageAttributeNames: []string{
			"All",
		},
	}

	resp, err := c.client.ReceiveMessage(receiveCtx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to receive messages: %w", err)
	}

	messages := make([]ports.SQSMessage, len(resp.Messages))
	for i, msg := range resp.Messages {
		messages[i] = ports.SQSMessage{
			MessageID:     aws.ToString(msg.MessageId),
			ReceiptHandle: aws.ToString(msg.ReceiptHandle),
			Body:          aws.ToString(msg.Body),
		}
	}

	return messages, nil
}

// DeleteMessage elimina un mensaje de la cola SQS
func (c *sqsClient) DeleteMessage(ctx context.Context, queueURL string, receiptHandle string) error {
	if queueURL == "" || receiptHandle == "" {
		return fmt.Errorf("queue URL and receipt handle cannot be empty")
	}

	deleteCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	input := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueURL),
		ReceiptHandle: aws.String(receiptHandle),
	}

	_, err := c.client.DeleteMessage(deleteCtx, input)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}
