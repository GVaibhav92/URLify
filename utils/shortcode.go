package utils

import (
	"crypto/rand"
	"math/big"
	"strings"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const defaultLength = 7

func GenerateShortCode() (string, error) {
	result := make([]byte, defaultLength)

	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}

	return string(result), nil
}

func ValidateCustomCode(code string) bool {
	if len(code) < 3 || len(code) > 30 {
		return false
	}
	for _, c := range code {
		if !strings.ContainsRune(charset, c) {
			return false
		}
	}
	return true
}
