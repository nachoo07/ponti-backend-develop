package pkglocalstack

import "time"

const (
	defaultTimeout = 30 * time.Second
	maxRetries     = 3
)

// Constantes para configuración de SQS
const (
	defaultWaitTimeSeconds   = 20
	defaultVisibilityTimeout = 30
	defaultMaxMessages       = 10
)
