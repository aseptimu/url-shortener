package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
)

type ConfigType struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseAddress   string `env:"BASE_URL"`
}

func NewConfig() *ConfigType {
	config := ConfigType{}
	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&config.BaseAddress, "b", "http://localhost:8080", "shorten URL base address")
	flag.Parse()

	if err := env.Parse(&config); err != nil {
		fmt.Printf("Ошибка загрузки конфигурации из env: %v\n", err)
	}

	return &config
}
