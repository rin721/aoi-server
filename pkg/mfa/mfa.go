package mfa

import (
	"strings"
	"time"

	"github.com/pquerna/otp/totp"
)

type TOTPKey struct {
	Secret string
	URL    string
}

func GenerateTOTP(issuer, accountName string) (TOTPKey, error) {
	key, err := totp.Generate(totp.GenerateOpts{Issuer: issuer, AccountName: accountName})
	if err != nil {
		return TOTPKey{}, err
	}
	return TOTPKey{Secret: key.Secret(), URL: key.URL()}, nil
}

func ValidateTOTP(code, secret string) bool {
	return totp.Validate(strings.TrimSpace(code), secret)
}

func GenerateTOTPCode(secret string, at time.Time) (string, error) {
	return totp.GenerateCode(secret, at)
}
