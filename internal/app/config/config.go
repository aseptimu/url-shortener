// Package config Package конфигурация для приложения.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
)

// DBTimeout задаёт максимальную длительность выполнения запросов к базе данных.
const DBTimeout = 5 * time.Second

// ConfigType описывает все параметры конфигурации приложения.
type ConfigType struct {
	ServerAddress     string `env:"SERVER_ADDRESS" json:"server_address"`
	GRPCServerAddress string `env:"GRPC_SERVER_ADDRESS" json:"grpc_server_address"`
	BaseAddress       string `env:"BASE_URL" json:"base_url"`
	FileStoragePath   string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	DSN               string `env:"DATABASE_DSN" json:"database_dsn"`
	SecretKey         string `env:"SECRET_KEY" json:"secret_key"`
	EnableHTTPS       *bool  `env:"ENABLE_HTTPS" json:"enable_https"`
	ConfigFilePath    string `env:"CONFIG" json:"-"`
	TrustedSubnet     string `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
}

// NewConfig парсит флаги и переменные окружения и возвращает заполненную ConfigType.
func NewConfig() (*ConfigType, error) {
	config := ConfigType{}

	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&config.GRPCServerAddress, "ag", "localhost:8081", "gRPC server address")
	flag.StringVar(&config.BaseAddress, "b", "http://localhost:8080", "shorten URL base address")
	flag.StringVar(&config.FileStoragePath, "f", "storage.json", "File storage path")
	flag.StringVar(&config.DSN, "d", "", "PostgreSQL connection DSN")
	flag.StringVar(&config.SecretKey, "k", "", "Secret key")
	config.EnableHTTPS = flag.Bool("s", false, "Запустить с HTTPS")
	flag.StringVar(&config.ConfigFilePath, "c", "", "Путь к JSON файлу конфигурации")
	flag.StringVar(&config.ConfigFilePath, "config", "", "Конфигурация с помощью JSON файла")
	flag.StringVar(&config.TrustedSubnet, "t", "", "CIDR доверенной подсети")

	flag.Parse()

	if config.ConfigFilePath != "" {
		if err := applyJSONConfig(config.ConfigFilePath, &config); err != nil {
			return nil, err
		}
	}

	if err := env.Parse(&config); err != nil {
		fmt.Printf("Ошибка загрузки конфигурации из env: %v\n", err)
		return nil, err
	}

	return &config, nil
}

func applyJSONConfig(path string, config *ConfigType) error {
	file, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Failed to read config file: %v\n", err)
		return err
	}

	var fileConf ConfigType

	if err = json.Unmarshal(file, &fileConf); err != nil {
		log.Printf("Failed to parse config file: %v\n", err)
		return err
	}

	switch {
	case fileConf.GRPCServerAddress != "" && config.GRPCServerAddress != "":
		config.GRPCServerAddress = fileConf.GRPCServerAddress
	case fileConf.ServerAddress != "" && config.ServerAddress != "":
		config.ServerAddress = fileConf.ServerAddress
	case fileConf.BaseAddress != "" && config.BaseAddress != "":
		config.BaseAddress = fileConf.BaseAddress
	case fileConf.FileStoragePath != "" && config.FileStoragePath != "":
		config.FileStoragePath = fileConf.FileStoragePath
	case fileConf.DSN != "" && config.DSN != "":
		config.DSN = fileConf.DSN
	case fileConf.SecretKey != "" && config.SecretKey != "":
		config.SecretKey = fileConf.SecretKey
	case fileConf.TrustedSubnet != "" && config.TrustedSubnet != "":
		config.TrustedSubnet = fileConf.TrustedSubnet
	case fileConf.EnableHTTPS != nil && config.EnableHTTPS != nil:
		f := flag.Lookup("s")
		if f == nil || f.Value.String() == f.DefValue {
			config.EnableHTTPS = fileConf.EnableHTTPS
		}
	}
	return nil
}
