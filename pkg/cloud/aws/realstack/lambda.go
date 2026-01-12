package pkgrealstack

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"

	"github.com/alphacodinggroup/ponti-backend/pkg/cloud/aws/ports"
)

// lambdaClient implementa la interfaz LambdaClient
type lambdaClient struct {
	client *lambda.Client
}

// NewLambdaClient crea una nueva instancia del cliente Lambda
func NewLambdaClient(cfg aws.Config) ports.LambdaClient {
	return &lambdaClient{
		client: lambda.NewFromConfig(cfg),
	}
}

// HandleRequest procesa una petición de API Gateway a través de Lambda
func (c *lambdaClient) HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	if err := c.validateRequest(&request); err != nil {
		return c.createErrorResponse(err, 400)
	}

	return c.processRequest(ctx, &request)
}

func (c *lambdaClient) validateRequest(request *events.APIGatewayProxyRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if len(request.Body) > maxPayloadSize {
		return fmt.Errorf("request body exceeds maximum size of %d bytes", maxPayloadSize)
	}

	switch request.HTTPMethod {
	case "GET", "POST", "PUT", "DELETE", "PATCH":
		return nil
	default:
		return fmt.Errorf("unsupported HTTP method: %s", request.HTTPMethod)
	}
}

func (c *lambdaClient) processRequest(ctx context.Context, request *events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	functionName := c.getFunctionName(request)
	return c.invokeLambda(ctx, functionName, request)
}

func (c *lambdaClient) invokeLambda(ctx context.Context, functionName string, request *events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	input := &lambda.InvokeInput{
		FunctionName:   aws.String(functionName),
		Payload:        []byte(request.Body),
		InvocationType: types.InvocationTypeRequestResponse,
	}

	result, err := c.client.Invoke(ctx, input)
	if err != nil {
		return c.createErrorResponse(err, 500)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: int(result.StatusCode),
		Body:       string(result.Payload),
		Headers:    c.getDefaultHeaders(),
	}, nil
}

func (c *lambdaClient) createErrorResponse(err error, statusCode int) (events.APIGatewayProxyResponse, error) {
	errorResponse := map[string]string{
		"error": err.Error(),
	}

	body, _ := json.Marshal(errorResponse)

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       string(body),
		Headers:    c.getDefaultHeaders(),
	}, nil
}

func (c *lambdaClient) getDefaultHeaders() map[string]string {
	return map[string]string{
		"Content-Type":                 "application/json",
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Headers": "Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token",
		"Access-Control-Allow-Methods": "GET,POST,PUT,DELETE,PATCH,OPTIONS",
	}
}

func (c *lambdaClient) getFunctionName(request *events.APIGatewayProxyRequest) string {
	return fmt.Sprintf("my-function-%s", request.HTTPMethod)
}
