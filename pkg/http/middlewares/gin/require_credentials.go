package pkgmwr

import (
	"net/http"

	"github.com/gin-gonic/gin"

	pkgtypes "github.com/alphacodinggroup/ponti-backend/pkg/types"
)

// RequireCredentials valida el payload de login para la autenticación del usuario.
func RequireCredentials() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var credentials pkgtypes.LoginCredentials

		if err := ctx.ShouldBindJSON(&credentials); err != nil {
			apiErr := pkgtypes.NewError(pkgtypes.ErrValidation, "Invalid request payload", err)
			ctx.AbortWithStatusJSON(http.StatusBadRequest, apiErr.ToJSON())
			return
		}
		if credentials.Username == "" && credentials.Email == "" {
			apiErr := pkgtypes.NewMissingFieldError("username/email")
			ctx.AbortWithStatusJSON(http.StatusBadRequest, apiErr.ToJSON())
			return
		}

		ctx.Set(ContextCredentials, credentials)
		ctx.Next()
	}
}
