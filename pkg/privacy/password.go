package privacy

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func HashPassword(password string, salt string) string {
	salted := salt + password
	hash := sha256.Sum256([]byte(salted))
	return hex.EncodeToString(hash[:])
}

func VerifyPassword(password, hash, salt string) bool {
	salted := salt + password
	computed := sha256.Sum256([]byte(salted))
	computedHex := hex.EncodeToString(computed[:])
	return strings.EqualFold(computedHex, hash)
}
