package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
)

func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		64*1024, 1, 4,
		hex.EncodeToString(salt),
		hex.EncodeToString(hash)), nil
}

func VerifyPassword(password, encodedHash string) bool {
	var version, memory, iterations, parallelism int
	var saltHex, hashHex string

	_, err := fmt.Sscanf(encodedHash, "$argon2id$v=%d$m=%d,t=%d,p=%d$", &version, &memory, &iterations, &parallelism)
	if err != nil {
		return false
	}

	prefix := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$", version, memory, iterations, parallelism)
	parts := strings.TrimPrefix(encodedHash, prefix)

	partsSplit := strings.SplitN(parts, "$", 2)
	if len(partsSplit) != 2 {
		return false
	}
	saltHex = partsSplit[0]
	hashHex = partsSplit[1]

	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		return false
	}

	hash, err := hex.DecodeString(hashHex)
	if err != nil {
		return false
	}

	computedHash := argon2.IDKey([]byte(password), salt, uint32(iterations), uint32(memory), uint8(parallelism), 32)

	return subtle.ConstantTimeCompare(hash, computedHash) == 1
}

func GenerateSessionToken() (string, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return hex.EncodeToString(token), nil
}

func SessionExpiry(hours int) time.Time {
	return time.Now().Add(time.Duration(hours) * time.Hour)
}
