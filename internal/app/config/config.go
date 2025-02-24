package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
)

type ConfigType struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseAddress     string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

func NewConfig() *ConfigType {
	config := ConfigType{}

	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&config.BaseAddress, "b", "http://localhost:8080", "shorten URL base address")
	flag.StringVar(&config.FileStoragePath, "f", "storage.json", "File storage path")

	flag.Parse()

	if err := env.Parse(&config); err != nil {
		fmt.Printf("Ошибка загрузки конфигурации из env: %v\n", err)
	}

	return &config
}
