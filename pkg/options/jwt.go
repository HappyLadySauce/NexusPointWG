package options

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
)

type JWTOptions struct {
	Secret     string        `json:"secret" mapstructure:"secret"`
	Expiration time.Duration `json:"expiration" mapstructure:"expiration"`
}

func NewJWTOptions() *JWTOptions {
	return &JWTOptions{
		Secret:     "your-secret-key-change-in-production",
		Expiration: 7 * 24 * time.Hour, // 7 days
	}
}

func (j *JWTOptions) Validate() []error {
	var errors []error
	if j.Secret == "" {
		errors = append(errors, fmt.Errorf("jwt secret is required"))
	}
	if j.Expiration <= 0 {
		errors = append(errors, fmt.Errorf("jwt expiration must be greater than 0"))
	}
	return errors
}

func (j *JWTOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&j.Secret, "jwt.secret", j.Secret, "JWT secret key used to sign tokens")
	fs.DurationVar(&j.Expiration, "jwt.expiration", j.Expiration, "JWT token expiration duration (e.g., 24h, 7d)")
}
