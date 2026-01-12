package pkgmwr

import (
	"bytes"
	"io"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-micro.dev/v4/logger"
)

type HttpLoggingOptions struct {
	LogLevel       string
	IncludeHeaders bool
	IncludeBody    bool
	ExcludedPaths  []string
}

// INFO: registra y loggea las solicitudes HTTP entrantes y las respuestas salientes
func RequestAndResponseLogger(options HttpLoggingOptions) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("Request and Response Logger Middleware: Starting...")
		// Generar un ID único para la solicitud
		requestID := uuid.New().String()
		c.Set("RequestID", requestID)

		// Verificar si la ruta está excluida
		for _, path := range options.ExcludedPaths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}

		// Registrar la solicitud entrante con el RequestID
		startTime := time.Now()
		logger.Infof("[%s] Incoming request: %s %s", requestID, c.Request.Method, c.Request.URL.Path)

		if options.IncludeHeaders {
			headers := make(map[string][]string)
			for k, v := range c.Request.Header {
				// Ejemplo: Omite headers que contienen información sensible
				if k != "Authorization" && k != "Cookie" {
					headers[k] = v
				}
			}
			logger.Infof("[%s] Request headers: %v", requestID, headers)
		}

		if options.IncludeBody {
			// Leer el cuerpo de la solicitud
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				// Registrar el cuerpo
				logger.Infof("[%s] Request body: %s", requestID, string(bodyBytes))
				// Restaurar el cuerpo para que los handlers posteriores puedan leerlo
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			} else {
				logger.Errorf("[%s] Failed to read request body: %v", requestID, err)
			}
		}

		// Procesar la solicitud
		c.Next()

		// Registrar la respuesta saliente
		endTime := time.Now()
		latency := endTime.Sub(startTime)
		statusCode := c.Writer.Status()
		logger.Infof("[%s] Response: %d, Latency: %v", requestID, statusCode, latency)
	}
}
