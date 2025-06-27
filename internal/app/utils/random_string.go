// Package utils содержит вспомогательные функции,
// в том числе для генерации случайных строк.
package utils

import (
	"math/rand"
	"time"
)

var seed = rand.New(rand.NewSource(time.Now().UnixNano()))

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// StringWithCharset возвращает строку длины length,
// символы берутся случайно из заданного charset.
func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seed.Intn(len(charset))]
	}
	return string(b)
}

// RandomString возвращает случайную строку длины length,
// используя предопределённый набор символов charset.
func RandomString(length int) string {
	return StringWithCharset(length, charset)
}
