package config

import (
	"flag"
	"os"
)

type Config struct {
	ServAddr    string
	BaseAddr    string
	LogLevel    string
	EnvLogLevel string
}

func GetConfig() *Config {
	cf := new(Config)
	flag.StringVar(&cf.ServAddr, "a", ":8080", "default HTTP-server addres")
	flag.StringVar(&cf.BaseAddr, "b", "http://localhost:8080", "base host addr for short URL response")
	flag.StringVar(&cf.LogLevel, "l", "Info", "log level")
	flag.Parse()

	if EnvServAddr := os.Getenv("SERVER_ADDRESS"); EnvServAddr != "" {
		cf.ServAddr = EnvServAddr
	}

	if EnvBaseAddr := os.Getenv("BASE_URL"); EnvBaseAddr != "" {
		cf.BaseAddr = EnvBaseAddr
	}

	if EnvLogLevel := os.Getenv("LOG_LEVEL"); EnvLogLevel != "" {
		cf.LogLevel = EnvLogLevel
	}

	return cf
}
