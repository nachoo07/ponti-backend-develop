package pkgmwr

import (
	"net/http"

	"github.com/gin-gonic/gin"

	pkgtypes "github.com/alphacodinggroup/ponti-backend/pkg/types"
)

// Constants for error messages.
const (
	errInvalidPayload = "Invalid request payload"
	errMissingField   = "Either username or email is required"
)

// ValidateCredentials middleware to validate the login payload.
func ValidateCredentials() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Log message can be added here if needed.
		var credentials pkgtypes.LoginCredentials

		// Try binding the JSON payload to the struct.
		if err := ctx.ShouldBindJSON(&credentials); err != nil {
			apiErr := pkgtypes.NewError(pkgtypes.ErrValidation, errInvalidPayload, err)
			ctx.JSON(http.StatusBadRequest, apiErr.ToJSON())
			ctx.Abort()
			return
		}

		// Validate that at least one of the optional fields is present.
		if credentials.Username == "" && credentials.Email == "" {
			apiErr := pkgtypes.NewMissingFieldError("username/email")
			ctx.JSON(http.StatusBadRequest, apiErr.ToJSON())
			ctx.Abort()
			return
		}

		// Save the validated credentials in the context for later use.
		ctx.Set("credentials", credentials)
		ctx.Next()
	}
}
