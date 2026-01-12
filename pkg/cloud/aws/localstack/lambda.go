package pkglocalstack

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"

	"github.com/alphacodinggroup/ponti-backend/pkg/cloud/aws/ports"
)

// lambdaClient implementa la interfaz LambdaClient
type lambdaClient struct {
	client   *lambda.Client
	endpoint string
}

// NewLambdaClient crea una nueva instancia del cliente Lambda
func NewLambdaClient(cfg aws.Config, endpoint string) ports.LambdaClient {
	client := lambda.NewFromConfig(cfg, func(o *lambda.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	return &lambdaClient{
		client:   client,
		endpoint: endpoint,
	}
}

// HandleRequest procesa una petición de API Gateway
func (c *lambdaClient) HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	if err := c.validateRequest(&request); err != nil {
		return c.createErrorResponse(err, http.StatusBadRequest)
	}

	functionName, err := c.resolveFunctionName(&request)
	if err != nil {
		return c.createErrorResponse(err, http.StatusBadRequest)
	}

	response, err := c.invokeFunction(ctx, functionName, &request)
	if err != nil {
		return c.createErrorResponse(err, http.StatusInternalServerError)
	}

	return response, nil
}

func (c *lambdaClient) validateRequest(request *events.APIGatewayProxyRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	switch request.HTTPMethod {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch:
		return nil
	default:
		return fmt.Errorf("unsupported HTTP method: %s", request.HTTPMethod)
	}
}

func (c *lambdaClient) resolveFunctionName(request *events.APIGatewayProxyRequest) (string, error) {
	route := request.Resource
	if route == "" {
		route = request.Path
	}

	if route == "" {
		return "", fmt.Errorf("no route or path specified in request")
	}

	return fmt.Sprintf("localstack-%s-%s-function",
		sanitizePath(route),
		strings.ToLower(request.HTTPMethod)), nil
}

func (c *lambdaClient) invokeFunction(ctx context.Context, functionName string, request *events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt*2) * time.Second)
		}

		input := &lambda.InvokeInput{
			FunctionName:   aws.String(functionName),
			Payload:        []byte(request.Body),
			InvocationType: types.InvocationTypeRequestResponse,
		}

		result, err := c.client.Invoke(ctx, input)
		if err == nil {
			return c.processLambdaResponse(result)
		}
		lastErr = err
	}

	return events.APIGatewayProxyResponse{}, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

func (c *lambdaClient) processLambdaResponse(result *lambda.InvokeOutput) (events.APIGatewayProxyResponse, error) {
	if result.FunctionError != nil {
		return c.createErrorResponse(
			fmt.Errorf("lambda function error: %s", *result.FunctionError),
			http.StatusInternalServerError,
		)
	}

	var response events.APIGatewayProxyResponse
	if err := json.Unmarshal(result.Payload, &response); err != nil {
		return c.createErrorResponse(err, http.StatusInternalServerError)
	}

	// Asegurar que siempre tengamos headers
	if response.Headers == nil {
		response.Headers = make(map[string]string)
	}

	// Agregar headers por defecto
	for k, v := range c.getDefaultHeaders() {
		if _, exists := response.Headers[k]; !exists {
			response.Headers[k] = v
		}
	}

	return response, nil
}

func (c *lambdaClient) createErrorResponse(err error, statusCode int) (events.APIGatewayProxyResponse, error) {
	body, _ := json.Marshal(map[string]string{
		"error":   err.Error(),
		"message": "Error in Localstack Lambda execution",
	})

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       string(body),
		Headers:    c.getDefaultHeaders(),
	}, nil
}

func (c *lambdaClient) getDefaultHeaders() map[string]string {
	return map[string]string{
		"Content-Type":                     "application/json",
		"Access-Control-Allow-Origin":      "*",
		"Access-Control-Allow-Headers":     "Content-Type,X-Amz-Date,Authorization,X-Api-Key",
		"Access-Control-Allow-Methods":     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		"Access-Control-Allow-Credentials": "true",
		"X-Localstack-Implementation":      "true",
	}
}

func sanitizePath(path string) string {
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	path = strings.ReplaceAll(path, "/", "-")
	path = strings.ReplaceAll(path, " ", "-")
	return strings.ToLower(path)
}
