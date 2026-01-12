package pkglocalstack

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"

	"github.com/alphacodinggroup/ponti-backend/pkg/cloud/aws/ports"
)

type sqsClient struct {
	client   *sqs.Client
	endpoint string
}

func NewSQSClient(cfg aws.Config, endpoint string) ports.SQSClient {
	client := sqs.NewFromConfig(cfg, func(o *sqs.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	return &sqsClient{
		client:   client,
		endpoint: endpoint,
	}
}

func (c *sqsClient) GetOrCreateQueueURL(ctx context.Context, queueName string) (string, error) {
	if queueName == "" {
		return "", fmt.Errorf("queue name cannot be empty")
	}

	queueName = sanitizeQueueName(queueName)

	getCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	out, err := c.client.GetQueueUrl(getCtx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})
	if err == nil {
		return *out.QueueUrl, nil
	}

	createCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	attributes := map[string]string{
		"DelaySeconds":                  "0",
		"MaximumMessageSize":            "262144",
		"MessageRetentionPeriod":        "345600",
		"VisibilityTimeout":             fmt.Sprintf("%d", defaultVisibilityTimeout),
		"ReceiveMessageWaitTimeSeconds": fmt.Sprintf("%d", defaultWaitTimeSeconds),
	}

	createOut, err := c.client.CreateQueue(createCtx, &sqs.CreateQueueInput{
		QueueName:  aws.String(queueName),
		Attributes: attributes,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create queue %s in Localstack: %w", queueName, err)
	}

	return *createOut.QueueUrl, nil
}

func (c *sqsClient) SendMessage(ctx context.Context, queueURL string, messageBody string) error {
	if err := validateQueueURL(queueURL); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt*2) * time.Second)
		}

		input := &sqs.SendMessageInput{
			QueueUrl:    aws.String(queueURL),
			MessageBody: aws.String(messageBody),
			MessageAttributes: map[string]types.MessageAttributeValue{
				"Environment": {
					DataType:    aws.String("String"),
					StringValue: aws.String("localstack"),
				},
			},
		}

		_, err := c.client.SendMessage(ctx, input)
		if err == nil {
			return nil
		}
		lastErr = err
	}

	return fmt.Errorf("failed to send message after %d retries: %w", maxRetries, lastErr)
}

func (c *sqsClient) ReceiveMessages(ctx context.Context, queueURL string, maxMessages int32) ([]ports.SQSMessage, error) {
	if err := validateQueueURL(queueURL); err != nil {
		return nil, err
	}

	if maxMessages <= 0 || maxMessages > defaultMaxMessages {
		maxMessages = defaultMaxMessages
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
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

	resp, err := c.client.ReceiveMessage(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to receive messages from Localstack: %w", err)
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

func (c *sqsClient) DeleteMessage(ctx context.Context, queueURL string, receiptHandle string) error {
	if err := validateQueueURL(queueURL); err != nil {
		return err
	}

	if receiptHandle == "" {
		return fmt.Errorf("receipt handle cannot be empty")
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt*2) * time.Second)
		}

		input := &sqs.DeleteMessageInput{
			QueueUrl:      aws.String(queueURL),
			ReceiptHandle: aws.String(receiptHandle),
		}

		_, err := c.client.DeleteMessage(ctx, input)
		if err == nil {
			return nil
		}
		lastErr = err
	}

	return fmt.Errorf("failed to delete message after %d retries: %w", maxRetries, lastErr)
}

// Funciones auxiliares
func sanitizeQueueName(name string) string {
	name = strings.ToLower(name)
	return strings.ReplaceAll(name, " ", "-")
}

func validateQueueURL(queueURL string) error {
	if queueURL == "" {
		return fmt.Errorf("queue URL cannot be empty")
	}

	if !strings.Contains(queueURL, "/sqs") {
		return fmt.Errorf("invalid Localstack SQS queue URL format")
	}

	return nil
}
