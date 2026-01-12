package pkgmwr

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ValidateUserIDHeader verifica que el header 'X-User-ID' esté presente en la solicitud.
func ValidateUserIDHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "MISSING_USER_ID",
				"message": "The 'X-User-ID' header is required",
			})
			c.Abort() // Detiene la ejecución de la solicitud
			return
		}
		// Guarda el userID en el contexto para que los handlers puedan acceder
		c.Set("userID", userID)

		c.Next() // Continúa con el siguiente middleware o handler
	}
}
