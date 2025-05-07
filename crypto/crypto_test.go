package crypto

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	text := "admin"
	hash, _ := HashPassword(text)
	fmt.Println(hash)
}

func TestVerifyPassword(t *testing.T) {
	text := "123456"

	// Prefix + Cost + Salt + Hashed Text
	hash3 := "$2a$10$ygWrRwHCzg2GUpz0UK40kuWAGva121VkScpcdMNsDCih2U/bL2qYy"
	bMatched := VerifyPassword(text, hash3)
	assert.True(t, bMatched)

	bMatched = VerifyPassword(text, hash3)
	assert.True(t, bMatched)
}

func TestVerifyPasswordWithSalt_CorrectPassword(t *testing.T) {
	password := "securePassword"
	salt, _ := GenerateSalt(16)
	hashedPassword, _ := HashPasswordWithSalt(password, salt)

	result := VerifyPasswordWithSalt(password, salt, hashedPassword)
	assert.True(t, result, "Password verification should succeed with correct password and salt")
}

func TestJwtToken(t *testing.T) {
	const bearerWord string = "Bearer"
	token := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjowfQ.XgcKAAjHbA6o4sxxbEaMi05ingWvKdCNnyW9wowbJvs"
	auths := strings.SplitN(token, " ", 2)
	assert.Equal(t, len(auths), 2)
	assert.Equal(t, strings.EqualFold(auths[0], bearerWord), true, "JWT token is missing")
}
