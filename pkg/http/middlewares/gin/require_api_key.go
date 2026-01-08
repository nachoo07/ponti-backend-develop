package pkgmwr

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// RequireAPIKey ensures the request contains a valid API key in the header.
func RequireAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := strings.TrimSpace(c.GetHeader(HeaderAPIKey))
		expectedKey := os.Getenv(EnvAPIKey)
		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "MISSING_API_KEY",
				"message": "API key is required in '" + HeaderAPIKey + "' header.",
			})
			return
		}
		if expectedKey == "" || apiKey != expectedKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "INVALID_API_KEY",
				"message": "API key is invalid.",
			})
			return
		}
		c.Set(ContextAPIKey, apiKey)
		c.Next()
	}
}
