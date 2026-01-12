// pkgutils/validation.go
package pkgutils

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	pkgtypes "github.com/alphacodinggroup/ponti-backend/pkg/types"
)

// ValidateRequest valida la solicitud y retorna un error enriquecido si falla.
// c: Contexto de Gin.
// req: Puntero a la estructura donde se deserializará la solicitud.
func ValidateRequest(c *gin.Context, req any) error {
	// Intentar parsear el JSON a la estructura proporcionada
	if err := c.ShouldBindJSON(req); err != nil {
		var validationErrors validator.ValidationErrors
		// Si el error es de validación, enriquecerlo con detalles del campo y la regla
		if errors.As(err, &validationErrors) {
			validationDetails := make(map[string]any)
			for _, fieldErr := range validationErrors {
				validationDetails[fieldErr.Field()] = map[string]string{
					"tag":     fieldErr.Tag(),
					"param":   fieldErr.Param(),
					"message": fieldErr.Error(),
				}
			}
			return pkgtypes.NewErrorWithContext(
				pkgtypes.ErrValidation,
				"Validation failed",
				err,
				map[string]any{
					"errors": validationDetails,
				},
			)
		}

		// Si no es un error de validación, manejarlo como un error de deserialización
		return pkgtypes.NewErrorWithContext(
			pkgtypes.ErrValidation,
			"Invalid payload format",
			err,
			map[string]any{
				"hint": "Ensure the JSON is well-formed and matches the expected structure",
			},
		)
	}

	return nil // No hay errores
}

// RespondWithError envía una respuesta HTTP estándar con el código de estado y el mensaje de error.
// c: Contexto de Gin.
// err: Error que contiene información enriquecida.
func RespondWithError(c *gin.Context, err error) {
	// Verificar si el error es del tipo `pkgtypes.Error`
	if enrichedError, ok := err.(*pkgtypes.Error); ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   enrichedError.Message, // Puedes ajustar esto según tu preferencia
			"details": enrichedError.Context,
		})
		return
	}

	// Responder con un mensaje genérico si el error no está enriquecido
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": "An unexpected error occurred",
	})
}
