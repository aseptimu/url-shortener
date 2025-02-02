package service

import (
	"errors"
	"net/url"

	"github.com/aseptimu/url-shortener/internal/app/utils"
)

var URLsMap = map[string]string{}

func isValidURL(input string) bool {
	parsedURI, err := url.ParseRequestURI(input)
	return err == nil && parsedURI.Scheme != "" && parsedURI.Host != ""
}

var ShortenURL = func(input string) (string, error) {
	if !isValidURL(input) {
		return "", errors.New("invalid URL format")
	}

	shortURL := utils.RandomString(6)

	URLsMap[shortURL] = input

	return shortURL, nil
}

var GetOriginalURL = func(input string) (string, bool) {
	originalURL, exists := URLsMap[input]
	return originalURL, exists
}
