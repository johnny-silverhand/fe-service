package utils

import (
	"crypto/sha256"
	"fmt"
	math "math/rand"
)

func HashSha256(text string) string {
	hash := sha256.New()
	hash.Write([]byte(text))

	return fmt.Sprintf("%x", hash.Sum(nil))
}

func HashDigit(n int) string {
	var digits = []rune("1234567890")
	b := make([]rune, n)
	for i := range b {
		b[i] = digits[math.Intn(len(digits))]
	}
	return string(b)
}
