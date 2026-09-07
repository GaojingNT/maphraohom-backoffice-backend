package encryption

import (
	"crypto/sha512"
	"fmt"
	"math/rand"
	"strings"
)

func EncryptPassword(password string, salt string) string {
	var (
		saltCharacters      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		saltLength          = 16
		saltBytes           = make([]byte, saltLength)
		hashedPassword      string
		combinePassword     string
		hashedFinalPassword string
	)

	// CASE: Encrypt payload password
	// 1. Generate salt
	// 2. Hash password + salt with SHA-512 algorithm
	// 3. Combine hashed password with salt
	// 4. Make hashed password with salt as string
	// 5. Make final hashed password

	// Generate salt
	if salt == "" {
		for i := range saltBytes {
			saltBytes[i] = saltCharacters[rand.Intn(len(saltCharacters))]
		}
		salt = string(saltBytes)
	} else {
		saltBytes = []byte(salt)
	}

	// Combine hashed password with salt
	combinePassword = password + string(salt)

	// Hash password + salt with SHA-512 algorithm
	hash := sha512.New()
	hash.Write([]byte(combinePassword))
	hash.Write(saltBytes)
	hashedPassword = fmt.Sprintf("%x", hash.Sum(nil))

	// Make final hashed password
	hashedFinalPassword = fmt.Sprintf("%s$%s", hashedPassword, salt)

	return hashedFinalPassword
}

func VerifyPassword(password string, hashedPassword string) bool {
	var (
		payloadHashedPassword string
		splitHashedPassword   []string
	)

	// CASE: Verify password
	// 1. Split hashed password with salt
	// 2. Encrypt payload password with original salt
	// 3. Compare payload hashed password with original hashed password
	// 4. Return result

	// Split hashed password with salt
	splitHashedPassword = strings.Split(hashedPassword, "$")

	// Encrypt payload password
	payloadHashedPassword = EncryptPassword(password, splitHashedPassword[1])

	// Compare hashed password with original hashed password
	return payloadHashedPassword == hashedPassword
}
