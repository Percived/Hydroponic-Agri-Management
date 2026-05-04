package response

import (
	"errors"
	"net/http"

	platformErrors "hydroponic-backend/internal/platform/errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func ValidationError(c *gin.Context, err error) {
	var vErrs validator.ValidationErrors
	errorsData := make([]gin.H, 0)
	if errors.As(err, &vErrs) {
		for _, fe := range vErrs {
			errorsData = append(errorsData, gin.H{
				"field":  fe.Field(),
				"reason": fe.Tag(),
			})
		}
	}
	Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "validation_error", gin.H{"errors": errorsData})
}
