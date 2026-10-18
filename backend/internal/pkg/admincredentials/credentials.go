// Package admincredentials shares first-install credential rules across setup and YAML bootstrap.
package admincredentials

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/mail"

	"github.com/gin-gonic/gin/binding"
)

func ValidateEmail(email string) bool {
	if _, err := mail.ParseAddress(email); err != nil || len(email) > 254 {
		return false
	}
	login := struct {
		Email string `binding:"required,email"`
	}{Email: email}
	return binding.Validator.ValidateStruct(&login) == nil
}

func ValidatePassword(password string) error {
	if len(password) < 8 || len(password) > 72 {
		return fmt.Errorf("password must be between 8 and 72 bytes")
	}
	return nil
}

func GenerateEmail() (string, error) {
	var suffix [6]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		return "", fmt.Errorf("generate admin email: %w", err)
	}
	return "admin-" + hex.EncodeToString(suffix[:]) + "@sub2api.local", nil
}
