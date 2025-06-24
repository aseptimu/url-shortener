// Package utils содержит вспомогательные функции.
package utils

import (
	"crypto/rand"
	"encoding/base64"
	"log"
)

const secretKeyLength = 32

// GenerateRandomSecretKey создаёт случайный байтовый массив длины secretKeyLength,
// заполняет его генератором rand,
// а затем возвращает его в виде base64 строки.
func GenerateRandomSecretKey() string {
	b := make([]byte, secretKeyLength)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal("Failed to generate random secret key:", err)
	}
	return base64.URLEncoding.EncodeToString(b)
}
