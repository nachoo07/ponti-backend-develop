package pkgmwr

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	pkgtypes "github.com/alphacodinggroup/ponti-backend/pkg/types"
)

// ErrorHandlingMiddleware captures errors added to the context and responds appropriately.
func ErrorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log the request method and URL for context.
		log.Printf("[ErrorHandlingMiddleware] Starting error handling for %s %s", c.Request.Method, c.Request.URL.Path)

		c.Next() // Process the request.

		// If a response has already been written, do not proceed.
		if c.Writer.Written() {
			return
		}

		// If there are errors in the context, process the first one.
		if len(c.Errors) > 0 {
			log.Printf("[ErrorHandlingMiddleware] Found %d error(s)", len(c.Errors))

			// Take the first error for the response.
			ginErr := c.Errors[0]
			log.Printf("[ErrorHandlingMiddleware] Error: %v", ginErr.Err)

			var status int
			var response any

			// Check if the error is a domain error (*pkgtypes.Error).
			var domainErr *pkgtypes.Error
			if errors.As(ginErr.Err, &domainErr) {
				apiErr, code := pkgtypes.NewAPIError(domainErr)
				response = apiErr.ToResponse()
				status = code
			} else {
				// Check if the error is already an API error (*pkgtypes.APIError).
				var apiErr *pkgtypes.APIError
				if errors.As(ginErr.Err, &apiErr) {
					response = apiErr.ToResponse()
					status = apiErr.Code
				} else {
					// For unknown errors, return an internal error with a generic message.
					response = gin.H{
						"error":   "INTERNAL_ERROR",
						"message": "An internal error occurred, please try again later.",
						"details": ginErr.Err.Error(),
					}
					status = http.StatusInternalServerError
				}
			}

			// Send a JSON response with the appropriate status code.
			c.JSON(status, response)
			// Abort further processing.
			c.Abort()
			return
		}
	}
}
