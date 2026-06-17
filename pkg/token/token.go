package token

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/google/uuid"
)

func GeneratePublicID() uuid.UUID {
	return uuid.New()
}

func GenerateQRCodeToken() string {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return uuid.NewString()
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}
