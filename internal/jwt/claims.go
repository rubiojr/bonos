package jwt

import (
	"context"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// Validate does nothing for this example.
func (c *Claims) Validate(ctx context.Context) error {
	return nil
}
