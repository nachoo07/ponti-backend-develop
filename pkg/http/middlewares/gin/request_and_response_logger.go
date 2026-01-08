package pkgmwr

import (
	"bytes"
	"io"
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

// RequestAndResponseLogger registra solicitudes y respuestas, omitiendo ExcludedPaths.
// INFO: registra y loggea las solicitudes HTTP entrantes y las respuestas salientes
func RequestAndResponseLogger(options HttpLoggingOptions) gin.HandlerFunc {
	excluded := make(map[string]struct{}, len(options.ExcludedPaths))
	for _, path := range options.ExcludedPaths {
		excluded[path] = struct{}{}
	}
	return func(c *gin.Context) {
		if _, skip := excluded[c.Request.URL.Path]; skip {
			c.Next()
			return
		}
		requestID := uuid.NewString()
		c.Set("RequestID", requestID)
		startTime := time.Now()
		logger.Infof("[%s] Incoming request: %s %s", requestID, c.Request.Method, c.Request.URL.Path)
		if options.IncludeHeaders {
			headers := make(map[string][]string)
			for k, v := range c.Request.Header {
				if k != "Authorization" && k != "Cookie" {
					headers[k] = v
				}
			}
			logger.Infof("[%s] Request headers: %v", requestID, headers)
		}
		if options.IncludeBody {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				logger.Infof("[%s] Request body: %s", requestID, string(bodyBytes))
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			} else {
				logger.Errorf("[%s] Failed to read request body: %v", requestID, err)
			}
		}
		c.Next()
		latency := time.Since(startTime)
		statusCode := c.Writer.Status()
		logger.Infof("[%s] Response: %d, Latency: %v", requestID, statusCode, latency)
	}
}
