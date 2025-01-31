package service

import (
	"errors"
	"github.com/aseptimu/url-shortener/internal/app/utils"
	"net/url"
)

var URLsMap = map[string]string{}

func isValidUrl(input string) bool {
	parsedUri, err := url.ParseRequestURI(input)
	return err == nil && parsedUri.Scheme != "" && parsedUri.Host != ""
}

func ShortenURL(input string) (string, error) {
	if !isValidUrl(input) {
		return "", errors.New("invalid URL format")
	}

	shortURL := utils.RandomString(6)

	URLsMap[shortURL] = input

	return shortURL, nil
}

func GetOriginalURL(input string) (string, bool) {
	originalURL, exists := URLsMap[input]
	return originalURL, exists
}
