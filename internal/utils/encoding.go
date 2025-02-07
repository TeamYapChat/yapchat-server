package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func GenerateToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)

	return hex.EncodeToString(bytes)
}

func HashUserID(id uint) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("verified-%d", id)))

	return hex.EncodeToString(hash[:])
}
