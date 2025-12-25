package validator

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin/binding"
	v10 "github.com/go-playground/validator/v10"
)

var (
	// urlSafeRegex matches only letters, numbers, underscores, and hyphens
	urlSafeRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

	// chineseRegex matches Chinese characters (CJK Unified Ideographs)
	// Using \p{Han} to match Han script characters (Chinese, Japanese Kanji, etc.)
	chineseRegex = regexp.MustCompile(`\p{Han}`)

	// allowedEmailDomains is a list of common email provider domains
	AllowedEmailDomains = []string{
		"qq.com",
		"163.com",
		"gmail.com",
		"outlook.com",
	}
)

func init() {
	// Register custom validators
	if v, ok := binding.Validator.Engine().(*v10.Validate); ok {
		if err := RegisterCustomValidators(v); err != nil {
			panic("Failed to register custom validators: " + err.Error())
		}
	}
}

// RegisterCustomValidators registers custom validation functions to the validator instance
func RegisterCustomValidators(v *v10.Validate) error {
	// Register urlsafe validator: only allows letters, numbers, underscores, and hyphens
	if err := v.RegisterValidation("urlsafe", validateURLSafe); err != nil {
		return err
	}

	// Register nochinese validator: disallows Chinese characters
	if err := v.RegisterValidation("nochinese", validateNoChinese); err != nil {
		return err
	}

	// Register emaildomain validator: only allows common email provider domains
	if err := v.RegisterValidation("emaildomain", validateEmailDomain); err != nil {
		return err
	}

	return nil
}

// validateURLSafe checks if the string contains only URL-safe characters
// (letters, numbers, underscores, and hyphens)
func validateURLSafe(fl v10.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // empty values are handled by required tag
	}
	return urlSafeRegex.MatchString(value)
}

// validateNoChinese checks if the string does not contain Chinese characters
func validateNoChinese(fl v10.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // empty values are handled by required tag
	}

	// Check for Chinese characters using regex
	return !chineseRegex.MatchString(value)
}

// validateEmailDomain checks if the email domain is in the allowed list
func validateEmailDomain(fl v10.FieldLevel) bool {
	email := fl.Field().String()
	if email == "" {
		return true // empty values are handled by required tag
	}

	// Parse email to extract domain
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false // invalid email format
	}

	domain := strings.ToLower(strings.TrimSpace(parts[1]))

	// Check if domain is in allowed list
	for _, allowedDomain := range AllowedEmailDomains {
		if domain == allowedDomain {
			return true
		}
	}

	return false
}

// ValidateURLSafeString is a helper function to validate a string directly
func ValidateURLSafeString(s string) bool {
	return urlSafeRegex.MatchString(s)
}

// ValidateNoChineseString is a helper function to validate a string has no Chinese
func ValidateNoChineseString(s string) bool {
	return !chineseRegex.MatchString(s)
}

// ValidateEmailDomainString is a helper function to validate an email domain
func ValidateEmailDomainString(email string) bool {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	domain := strings.ToLower(strings.TrimSpace(parts[1]))
	for _, allowedDomain := range AllowedEmailDomains {
		if domain == allowedDomain {
			return true
		}
	}
	return false
}

// IsURLSafe checks if a string can be safely used in a URL path segment
func IsURLSafe(s string) bool {
	// Try to encode as URL path segment
	encoded := url.PathEscape(s)
	// If encoding changes the string, it contains unsafe characters
	return encoded == s || len(encoded) == len(s)
}
