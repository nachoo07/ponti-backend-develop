package lot

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	pkgmwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/alphacodinggroup/ponti-backend/pkg/validations"
	dto "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/handler/dto"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
)

// ValidationError representa un error de validación específico
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// ValidationErrors es una colección de errores de validación
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// Error implementa la interfaz error
func (v ValidationErrors) Error() string {
	if len(v.Errors) == 0 {
		return "validation failed"
	}
	return fmt.Sprintf("validation failed: %s", v.Errors[0].Message)
}

// Funciones de validación específicas del negocio que fueron removidas del paquete core

// ValidateLotName valida nombres de lotes con reglas específicas del negocio
func ValidateLotName(name string) error {
	return ValidateNameExtended(name, "lot", 2, 255)
}

// ValidateFieldName valida nombres de campos con reglas específicas del negocio
func ValidateFieldName(name string) error {
	return ValidateNameExtended(name, "field", 2, 255)
}

// ValidateCropName valida nombres de cultivos con reglas específicas del negocio
func ValidateCropName(name string) error {
	return ValidateNameExtended(name, "crop", 2, 255)
}

// ValidateNameExtended valida nombres con soporte extendido de caracteres para entidades del negocio
func ValidateNameExtended(name, fieldName string, minLen, maxLen int) error {
	if strings.TrimSpace(name) == "" {
		return validations.Err(fieldName, "cannot be empty")
	}

	if len(strings.TrimSpace(name)) < minLen {
		return validations.Err(fieldName, fmt.Sprintf("must have at least %d characters", minLen))
	}

	if len(name) > maxLen {
		return validations.Err(fieldName, fmt.Sprintf("cannot exceed %d characters", maxLen))
	}

	if strings.Contains(name, "  ") {
		return validations.Err(fieldName, "cannot contain consecutive spaces")
	}

	// Validar que solo contenga caracteres válidos
	if !isValidBusinessName(name) {
		return validations.Err(fieldName, "contains invalid characters")
	}

	return nil
}

// isValidBusinessName valida nombres de negocio con soporte extendido de caracteres
func isValidBusinessName(name string) bool {
	// Permite letras, números, espacios, guiones, apóstrofes, puntos, guiones bajos, paréntesis, corchetes, llaves, ampersands
	pattern := `^[a-zA-ZáéíóúÁÉÍÓÚñÑ0-9\s\-'\._\(\)\[\]\{\}&]+$`
	matched, _ := regexp.MatchString(pattern, name)
	return matched
}

// ValidateNonEmptyCollection valida que una colección no esté vacía
func ValidateNonEmptyCollection(collection any, fieldName string) error {
	switch v := collection.(type) {
	case []any:
		if len(v) == 0 {
			return validations.Err(fieldName, "cannot be empty")
		}
	case []string:
		if len(v) == 0 {
			return validations.Err(fieldName, "cannot be empty")
		}
	case []int:
		if len(v) == 0 {
			return validations.Err(fieldName, "cannot be empty")
		}
	case []int64:
		if len(v) == 0 {
			return validations.Err(fieldName, "cannot be empty")
		}
	case []dto.LotDates:
		if len(v) == 0 {
			return validations.Err(fieldName, "cannot be empty")
		}
	default:
		return validations.Err(fieldName, "unsupported collection type")
	}
	return nil
}

// ValidateHectares valida valores de hectáreas
func ValidateHectares(hectares decimal.Decimal, fieldName string) error {
	if hectares.LessThanOrEqual(decimal.Zero) {
		return validations.Err(fieldName, "must be greater than 0")
	}

	// Opcional: Agregar validación de límite máximo
	maxHectares := decimal.NewFromFloat(10000.0) // 10,000 hectares as example limit
	if hectares.GreaterThan(maxHectares) {
		return validations.Err(fieldName, "exceeds maximum allowed hectares")
	}

	return nil
}

// ValidateTons valida valores de toneladas
func ValidateTons(tons decimal.Decimal, fieldName string) error {
	if tons.IsNegative() {
		return validations.Err(fieldName, "must be greater than or equal to 0")
	}

	// Límite máximo razonable de toneladas por lote
	maxTons := decimal.NewFromFloat(10000.0)
	if tons.GreaterThan(maxTons) {
		return validations.Err(fieldName, "exceeds maximum allowed tons")
	}

	return nil
}

// ValidateSeason valida el campo temporada
func ValidateSeason(season string, fieldName string) error {
	if strings.TrimSpace(season) == "" {
		return validations.Err(fieldName, "cannot be empty")
	}

	// Validar formato de temporada (ej: "2024-2025", "2025")
	seasonPattern := `^(\d{4}(-\d{4})?)$`
	matched, _ := regexp.MatchString(seasonPattern, season)
	if !matched {
		return validations.Err(fieldName, "invalid season format. Use format: YYYY or YYYY-YYYY")
	}

	return nil
}

// ValidateCropID valida IDs de cultivos
func ValidateCropID(cropID int64, fieldName string) error {
	if cropID <= 0 {
		return validations.Err(fieldName, "must be greater than 0")
	}
	return nil
}

// ValidateFieldID valida IDs de campos
func ValidateFieldID(fieldID int64, fieldName string) error {
	if fieldID <= 0 {
		return validations.Err(fieldName, "must be greater than 0")
	}
	return nil
}

// ValidateLot valida todos los campos del lote según las reglas del negocio
func ValidateLot(req *dto.Lot) *ValidationErrors {
	errors := &ValidationErrors{}

	// Validaciones del nombre del lote
	validateLotName(req.Name, errors)

	// Validaciones del ID del campo
	validateFieldID(req.FieldID, errors)

	// Validaciones de hectáreas
	validateHectares(req.Hectares, errors)

	// Validaciones de cultivos
	validateCropID(req.PreviousCropID, errors, "previous_crop_id")
	validateCropID(req.CurrentCropID, errors, "current_crop_id")

	// Validaciones de temporada
	validateSeason(req.Season, errors)

	// Validaciones de fechas (si se proporcionan)
	validateLotDates(req.Dates, errors)

	return errors
}

// validateLotName valida el campo Nombre del Lote
func validateLotName(name string, errors *ValidationErrors) {
	if err := ValidateLotName(name); err != nil {
		errors.Errors = append(errors.Errors, ValidationError{
			Field:   "name",
			Message: err.Error(),
			Value:   name,
		})
	}
}

// validateFieldID valida el campo ID del Campo
func validateFieldID(fieldID int64, errors *ValidationErrors) {
	if err := ValidateFieldID(fieldID, "field_id"); err != nil {
		errors.Errors = append(errors.Errors, ValidationError{
			Field:   "field_id",
			Message: err.Error(),
			Value:   fmt.Sprintf("%d", fieldID),
		})
	}
}

// validateHectares valida el campo Hectáreas
func validateHectares(hectares decimal.Decimal, errors *ValidationErrors) {
	if err := ValidateHectares(hectares, "hectares"); err != nil {
		errors.Errors = append(errors.Errors, ValidationError{
			Field:   "hectares",
			Message: err.Error(),
			Value:   hectares.String(),
		})
	}
}

// validateCropID valida IDs de cultivos
func validateCropID(cropID int64, errors *ValidationErrors, fieldName string) {
	if err := ValidateCropID(cropID, fieldName); err != nil {
		errors.Errors = append(errors.Errors, ValidationError{
			Field:   fieldName,
			Message: err.Error(),
			Value:   fmt.Sprintf("%d", cropID),
		})
	}
}

// validateSeason valida el campo Temporada
func validateSeason(season string, errors *ValidationErrors) {
	if err := ValidateSeason(season, "season"); err != nil {
		errors.Errors = append(errors.Errors, ValidationError{
			Field:   "season",
			Message: err.Error(),
			Value:   season,
		})
	}
}

// validateLotDates valida las fechas del lote
func validateLotDates(dates []dto.LotDates, errors *ValidationErrors) {
	// Las fechas son opcionales, pero si se proporcionan, deben ser válidas
	if len(dates) > 0 {
		for i, date := range dates {
			validateLotDate(date, errors, i)
		}
	}
}

// validateLotDate valida una fecha individual del lote
func validateLotDate(date dto.LotDates, errors *ValidationErrors, index int) {
	// Validar fecha de siembra si se proporciona
	if date.SowingDate != "" {
		if _, err := validations.ValidateISODate(fmt.Sprintf("dates[%d].sowing_date", index), date.SowingDate); err != nil {
			errors.Errors = append(errors.Errors, ValidationError{
				Field:   fmt.Sprintf("dates[%d].sowing_date", index),
				Message: err.Error(),
				Value:   date.SowingDate,
			})
		}
	}

	// Validar fecha de cosecha si se proporciona
	if date.HarvestDate != "" {
		if _, err := validations.ValidateISODate(fmt.Sprintf("dates[%d].harvest_date", index), date.HarvestDate); err != nil {
			errors.Errors = append(errors.Errors, ValidationError{
				Field:   fmt.Sprintf("dates[%d].harvest_date", index),
				Message: err.Error(),
				Value:   date.HarvestDate,
			})
		}
	}

	// Validar secuencia
	if date.Sequence <= 0 {
		errors.Errors = append(errors.Errors, ValidationError{
			Field:   fmt.Sprintf("dates[%d].sequence", index),
			Message: "sequence must be greater than 0",
			Value:   fmt.Sprintf("%d", date.Sequence),
		})
	}
}

// ValidateLotRequest es un middleware que valida requests de lotes
func ValidateLotRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.Lot

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, types.ErrorResponse{
				Error: fmt.Sprintf("Error in request format: %v", err),
			})
			c.Abort()
			return
		}

		// Aplicar validaciones personalizadas
		if validationErrors := ValidateLot(&req); len(validationErrors.Errors) > 0 {
			c.JSON(http.StatusBadRequest, validationErrors)
			c.Abort()
			return
		}

		// Establecer ID de usuario desde el contexto
		userID := c.Value(pkgmwr.ContextUserID)
		if userID != nil {
			if s, ok := userID.(string); ok {
				if i, err := strconv.ParseInt(s, 10, 64); err == nil {
					req.CreatedBy = &i
					req.UpdatedBy = &i
				}
			}
		}

		// Validar los campos del Base
		now := time.Now()
		base := &shareddomain.Base{
			CreatedAt: now,
			UpdatedAt: now,
		}

		// Validar el Base
		validationErrors := &ValidationErrors{}
		validateLotBase(base, validationErrors)

		// Si hay errores de validación del Base, abortar
		if len(validationErrors.Errors) > 0 {
			c.JSON(http.StatusBadRequest, validationErrors)
			c.Abort()
			return
		}

		// Si todas las validaciones pasan, continuar
		c.Set("validated_lot", &req)
		c.Next()
	}
}

// ValidateLotUpdate es un middleware que valida actualizaciones de lotes
func ValidateLotUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.LotUpdate

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, types.ErrorResponse{
				Error: fmt.Sprintf("Error in request format: %v", err),
			})
			c.Abort()
			return
		}

		// // Para actualizaciones, validar solo campos que estén presentes
		// validationErrors := &ValidationErrors{}

		// // Solo validar campos que no estén vacíos
		// if req.Name != "" {
		// 	validateLotName(req.Name, validationErrors)
		// }

		// if req.FieldID > 0 {
		// 	validateFieldID(req.FieldID, validationErrors)
		// }

		// if req.Hectares.GreaterThan(decimal.Zero) {
		// 	validateHectares(req.Hectares, validationErrors)
		// }

		// if req.PreviousCropID > 0 {
		// 	validateCropID(req.PreviousCropID, validationErrors, "previous_crop_id")
		// }

		// if req.CurrentCropID > 0 {
		// 	validateCropID(req.CurrentCropID, validationErrors, "current_crop_id")
		// }

		// // Para actualizaciones, las fechas son opcionales
		// if len(req.Dates) > 0 {
		// 	validateLotDates(req.Dates, validationErrors)
		// }

		// if len(validationErrors.Errors) > 0 {
		// 	c.JSON(http.StatusBadRequest, validationErrors)
		// 	c.Abort()
		// 	return
		// }

		// Set user ID from context
		userID := c.Value(pkgmwr.ContextUserID)
		if userID != nil {
			if s, ok := userID.(string); ok {
				if i, err := strconv.ParseInt(s, 10, 64); err == nil {
					req.UpdatedBy = &i
				}
			}
		}

		// Establecer timestamp actual para UpdatedAt
		now := time.Now()
		req.UpdatedAt = now

		// Si hay errores de validación del Base, abortar
		// if len(validationErrors.Errors) > 0 {
		// 	c.JSON(http.StatusBadRequest, validationErrors)
		// 	c.Abort()
		// 	return
		// }

		c.Set("validated_lot", &req)
		c.Next()
	}
}

// ValidateLotTonsUpdate es un middleware que valida actualizaciones de toneladas
func ValidateLotTonsUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Tons string `json:"tons" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, types.ErrorResponse{
				Error: fmt.Sprintf("Error in request format: %v", err),
			})
			c.Abort()
			return
		}

		// Validar formato de toneladas
		tons, err := decimal.NewFromString(req.Tons)
		if err != nil {
			c.JSON(http.StatusBadRequest, types.ErrorResponse{
				Error: "Invalid tons format. Must be a valid decimal number.",
			})
			c.Abort()
			return
		}

		// Validar valor de toneladas
		if err := ValidateTons(tons, "tons"); err != nil {
			c.JSON(http.StatusBadRequest, types.ErrorResponse{
				Error: err.Error(),
			})
			c.Abort()
			return
		}

		// Si todas las validaciones pasan, continuar
		c.Set("validated_tons", tons)
		c.Next()
	}
}

// validateLotBase valida los campos del Base (CreatedAt, UpdatedAt, Version)
func validateLotBase(base *shareddomain.Base, errors *ValidationErrors) {
	// Validar que CreatedAt no sea en el futuro
	createdAtStr := base.CreatedAt.Format("2006-01-02")
	if _, err := validations.ValidateISODate("created_at", createdAtStr); err != nil {
		errors.Errors = append(errors.Errors, ValidationError{
			Field:   "created_at",
			Message: err.Error(),
			Value:   createdAtStr,
		})
	}

	// Validar que UpdatedAt no sea en el futuro
	updatedAtStr := base.UpdatedAt.Format("2006-01-02")
	if _, err := validations.ValidateISODate("updated_at", updatedAtStr); err != nil {
		errors.Errors = append(errors.Errors, ValidationError{
			Field:   "updated_at",
			Message: err.Error(),
			Value:   updatedAtStr,
		})
	}

	// Validar que UpdatedAt no sea anterior a CreatedAt
	if base.UpdatedAt.Before(base.CreatedAt) {
		errors.Errors = append(errors.Errors, ValidationError{
			Field:   "updated_at",
			Message: "updated_at cannot be before created_at",
			Value:   updatedAtStr,
		})
	}
}
