package config

import (
	"errors"
	"os"
)

func HMACSecret() string {
	return os.Getenv("BONOS_HMAC_SECRET")
}

func Validate() error {
	if HMACSecret() == "" {
		return errors.New("missing BONOS_HMAC_SECRET env")
	}

	return nil
}
