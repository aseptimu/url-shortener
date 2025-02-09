package config

import "flag"

type ConfigType struct {
	ServerAddress string
	BaseAddress   string
}

var Config ConfigType

func init() {
	flag.StringVar(&Config.ServerAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&Config.BaseAddress, "b", "http://localhost:8080", "shorten URL base address")
}
