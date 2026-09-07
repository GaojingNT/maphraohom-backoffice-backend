package helpers

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"
)

func RandomHexString(length int) (string, error) {
	bytes := make([]byte, length)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func IncrementStringNumber(number string) (string, error) {
	if number == "" {
		return "1", nil
	}

	// Convert string to int
	n, err := strconv.Atoi(number)
	if err != nil {
		return "", err
	}

	return strconv.Itoa(n + 1), nil
}
