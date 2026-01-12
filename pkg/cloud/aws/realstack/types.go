package pkgrealstack

import "time"

const (
	defaultTimeout = 30 * time.Second
	maxPayloadSize = 6 * 1024 * 1024 // 6MB límite de AWS Lambda
)

// Constantes para configuración de SQS
const (
	defaultWaitTimeSeconds     = 20
	defaultVisibilityTimeout   = 30
	defaultMaxNumberOfMessages = 10
)
