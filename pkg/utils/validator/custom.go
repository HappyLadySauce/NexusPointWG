package validator

import (
	"net/netip"
	"net/url"
	"regexp"
	"strconv"
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

	// Register cidr validator: validates CIDR format (supports comma-separated multiple CIDRs)
	if err := v.RegisterValidation("cidr", validateCIDR); err != nil {
		return err
	}

	// Register endpoint validator: validates Endpoint format (host:port)
	if err := v.RegisterValidation("endpoint", validateEndpoint); err != nil {
		return err
	}

	// Register ipv4 validator: validates IPv4 address format (without CIDR)
	if err := v.RegisterValidation("ipv4", validateIPv4); err != nil {
		return err
	}

	// Register dnslist validator: validates comma-separated IP addresses (IPv4 or IPv6)
	if err := v.RegisterValidation("dnslist", validateDNSList); err != nil {
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

// validateCIDR 验证 CIDR 格式（支持逗号分隔的多个 CIDR）
func validateCIDR(fl v10.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // empty values are handled by required tag
	}

	// 支持逗号分隔的多个 CIDR
	parts := strings.Split(value, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		_, err := netip.ParsePrefix(part)
		if err != nil {
			return false
		}
	}
	return true
}

// validateEndpoint 验证 Endpoint 格式（host:port）
func validateEndpoint(fl v10.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // empty values are handled by required tag
	}

	// 格式：host:port
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return false
	}

	host := strings.TrimSpace(parts[0])
	port := strings.TrimSpace(parts[1])

	if host == "" || port == "" {
		return false
	}

	// 验证端口号
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return false
	}
	if portNum < 1 || portNum > 65535 {
		return false
	}

	return true
}

// validateIPv4 validates that the string is a valid IPv4 address (without CIDR notation).
func validateIPv4(fl v10.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // empty values are handled by required tag
	}

	// Parse as IP address
	ip, err := netip.ParseAddr(value)
	if err != nil {
		return false
	}

	// Check if it's IPv4
	return ip.Is4()
}

// validateDNSList validates that the string is a comma-separated list of valid IP addresses (IPv4 or IPv6).
func validateDNSList(fl v10.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // empty values are handled by required tag
	}

	// Support comma-separated multiple IP addresses
	parts := strings.Split(value, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue // Skip empty parts
		}
		// Parse as IP address (IPv4 or IPv6)
		_, err := netip.ParseAddr(part)
		if err != nil {
			return false
		}
	}
	return true
}
