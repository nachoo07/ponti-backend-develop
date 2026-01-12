package ports

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
)

// Provider types
const (
	ProviderAWS        = "aws"
	ProviderLocalstack = "localstack"
)

// Available AWS Services
const (
	ServiceS3             = "s3"
	ServiceSQS            = "sqs"
	ServiceRDS            = "rds"
	ServiceLambda         = "lambda"
	ServiceECS            = "ecs"
	ServiceSecretsManager = "secretsmanager"
)

// ValidServices define los servicios AWS soportados
var ValidServices = map[string]bool{
	ServiceS3:             true,
	ServiceSQS:            true,
	ServiceRDS:            true,
	ServiceLambda:         true,
	ServiceECS:            true,
	ServiceSecretsManager: true,
}

// Stack define la interfaz principal para todos los proveedores AWS
type Stack interface {
	Connect() error
	GetConfig() aws.Config
	NewSQSClient() SQSClient
	NewLambdaClient() LambdaClient
}

// Config define la configuración común para todos los proveedores
type Config interface {
	GetProvider() string
	GetAwsAccessKeyID() string
	GetAwsSecretAccessKey() string
	GetAwsRegion() string
	GetEndpoint() string
	SetEndpoint(string)
	GetServices() []string
	SetServices([]string)
	Validate() error
}

// SQSClient define las operaciones disponibles para SQS
type SQSClient interface {
	GetOrCreateQueueURL(ctx context.Context, queueName string) (string, error)
	SendMessage(ctx context.Context, queueURL, messageBody string) error
	ReceiveMessages(ctx context.Context, queueURL string, maxMessages int32) ([]SQSMessage, error)
	DeleteMessage(ctx context.Context, queueURL, receiptHandle string) error
}

// LambdaClient define las operaciones disponibles para Lambda
type LambdaClient interface {
	HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)
}

// SQSMessage define la estructura de un mensaje SQS
type SQSMessage struct {
	MessageID     string
	ReceiptHandle string
	Body          string
}
