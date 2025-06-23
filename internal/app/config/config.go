// Package config Package конфигурация для приложения.
package config

import (
	"flag"
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

// DBTimeout задаёт максимальную длительность выполнения запросов к базе данных.
const DBTimeout = 5 * time.Second

// ConfigType описывает все параметры конфигурации приложения.
type ConfigType struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseAddress     string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DSN             string `env:"DATABASE_DSN"`
	SecretKey       string `env:"SECRET_KEY"`
}

// NewConfig парсит флаги и переменные окружения и возвращает заполненную ConfigType.
func NewConfig() *ConfigType {
	config := ConfigType{}

	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&config.BaseAddress, "b", "http://localhost:8080", "shorten URL base address")
	flag.StringVar(&config.FileStoragePath, "f", "storage.json", "File storage path")
	flag.StringVar(&config.DSN, "d", "", "PostgreSQL connection DSN")
	flag.StringVar(&config.SecretKey, "s", "", "Secret key")

	flag.Parse()

	if err := env.Parse(&config); err != nil {
		fmt.Printf("Ошибка загрузки конфигурации из env: %v\n", err)
	}

	return &config
}
