package pkgmwr

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// RequireUserIDHeader asegura que un header de ID de usuario válido esté presente.
func RequireUserIDHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := strings.TrimSpace(c.GetHeader(HeaderUserID))
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":   "MISSING_USER_ID",
				"message": "User ID is required in '" + HeaderUserID + "' header.",
			})
			return
		}
		c.Set(ContextUserID, userID)
		// Propagar también al request.Context() para capas que usan ctx estándar
		ctx := context.WithValue(c.Request.Context(), ContextUserID, userID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
