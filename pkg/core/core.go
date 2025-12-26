package core

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/HappyLadySauce/errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"k8s.io/klog/v2"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	customvalidator "github.com/HappyLadySauce/NexusPointWG/pkg/utils/validator"
)

// ErrResponse defines the return messages when an error occurred.
// Reference will be omitted if it does not exist.
// swagger:model
type ErrResponse struct {
	// Code defines the business error code.
	Code int `json:"code"`

	// Message contains the detail of this message.
	// This message is suitable to be exposed to external
	Message string `json:"message"`

	// Reference returns the reference document which maybe useful to solve this error.
	Reference string `json:"reference,omitempty"`

	// Details contains detailed validation error information.
	// This field is only present when validation errors occur.
	Details map[string]string `json:"details,omitempty"`
}

// SuccessResponse defines the return messages when an operation succeeds.
// swagger:model
type SuccessResponse struct {
	// Code defines the business success code.
	Code int `json:"code"`

	// Message contains the success message.
	Message string `json:"message"`
}

// FormatValidationError formats validator.ValidationErrors into a user-friendly map.
// Returns a map where keys are field names (JSON field names) and values are error messages.
func FormatValidationError(err error) map[string]string {
	details := make(map[string]string)

	// Check if the error is a validator.ValidationErrors
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			// Get the field name (remove struct name prefix if present, e.g., "User.Password" -> "Password")
			fieldName := fieldError.Field()
			if idx := strings.Index(fieldName, "."); idx != -1 {
				fieldName = fieldName[idx+1:]
			}

			// Convert struct field name to JSON field name (camelCase to lowercase)
			// This matches the JSON tag naming convention used in this project
			jsonFieldName := strings.ToLower(fieldName[:1]) + fieldName[1:]

			// Generate user-friendly error message based on the validation tag
			var message string
			switch fieldError.Tag() {
			case "required":
				message = fmt.Sprintf("%s is required", jsonFieldName)
			case "min":
				message = fmt.Sprintf("%s must be at least %s characters", jsonFieldName, fieldError.Param())
			case "max":
				message = fmt.Sprintf("%s must be at most %s characters", jsonFieldName, fieldError.Param())
			case "email":
				message = fmt.Sprintf("%s must be a valid email address", jsonFieldName)
			case "len":
				message = fmt.Sprintf("%s must be exactly %s characters", jsonFieldName, fieldError.Param())
			case "oneof":
				message = fmt.Sprintf("%s must be one of: %s", jsonFieldName, fieldError.Param())
			case "urlsafe":
				message = fmt.Sprintf("%s must contain only letters, numbers, underscores, and hyphens", jsonFieldName)
			case "nochinese":
				message = fmt.Sprintf("%s must not contain Chinese characters", jsonFieldName)
			case "emaildomain":
				message = fmt.Sprintf("%s must use a common email provider domain (%s)", jsonFieldName, strings.Join(customvalidator.AllowedEmailDomains, ", "))
			default:
				message = fmt.Sprintf("%s failed validation on tag '%s'", jsonFieldName, fieldError.Tag())
			}

			details[jsonFieldName] = message
		}
	} else {
		// If it's not a ValidationErrors, just use the error message
		details["error"] = err.Error()
	}

	return details
}

// WriteResponseWithDetails write an error or the response data into http response body with details.
// The details parameter can be used to provide additional error information, such as validation errors.
func WriteResponseWithDetails(c *gin.Context, err error, data interface{}, details map[string]string) {
	if err != nil {
		klog.Errorf("%#+v", err)
		coder := errors.ParseCoder(err)

		response := ErrResponse{
			Code:      coder.Code(),
			Message:   coder.String(),
			Reference: coder.Reference(),
		}

		// Only include details if there are validation errors
		if len(details) > 0 {
			response.Details = details
		}

		c.JSON(coder.HTTPStatus(), response)
		return
	}

	// If data is nil, return a default success response
	if data == nil {
		response := SuccessResponse{
			Code:    code.ErrSuccess,
			Message: code.Message(code.ErrSuccess),
		}
		c.JSON(http.StatusOK, response)
		return
	}

	c.JSON(http.StatusOK, data)
}

// WriteResponse write an error or the response data into http response body.
// It use errors.ParseCoder to parse any error into errors.Coder
// errors.Coder contains error code, user-safe error message and http status code.
func WriteResponse(c *gin.Context, err error, data interface{}) {
	WriteResponseWithDetails(c, err, data, nil)
}

// WriteResponseBindErr writes a validation error response with details.
// It use FormatValidationError to format the validation errors into a map.
// Then it use WriteResponseWithDetails to write the response with the details.
func WriteResponseBindErr(c *gin.Context, err error, data interface{}) {
	details := FormatValidationError(err)
	WriteResponseWithDetails(c, errors.WithCode(code.ErrBind, "Validation failed"), data, details)
}
