package config

import (
	"flag"
	"log"
	"os"
)

type Config struct {
	ServAddr      string
	BaseAddr      string
	LogLevel      string
	FileStorePath string
	DBdsn         string
	StoreType     int
}

func GetConfig() *Config {
	cf := new(Config)
	flag.StringVar(&cf.ServAddr, "a", ":8080", "default HTTP-server addres")
	flag.StringVar(&cf.BaseAddr, "b", "http://localhost:8080", "base host addr for short URL response")
	flag.StringVar(&cf.LogLevel, "l", "Info", "log level")
	flag.StringVar(&cf.FileStorePath, "f", "/tmp/short-url-db.json", "full file name for storing url info")
	//query := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
	//	`localhost`, `5432`, `user`, `user_pass`, `urlstore`)
	flag.StringVar(&cf.DBdsn, "d" /*"***postgres:5432/praktikum?sslmode=disable" */ /*query */, "", "dsn for db connection")
	flag.Parse()
	log.Println("flag a: " + cf.ServAddr)
	log.Println("flag b: " + cf.BaseAddr)

	log.Println("flag l: " + cf.LogLevel)
	log.Println("flag f: " + cf.FileStorePath)
	log.Println("flag d: " + cf.DBdsn)
	if EnvServAddr := os.Getenv("SERVER_ADDRESS"); EnvServAddr != "" {
		cf.ServAddr = EnvServAddr
	}

	if EnvBaseAddr := os.Getenv("BASE_URL"); EnvBaseAddr != "" {
		cf.BaseAddr = EnvBaseAddr
	}

	if EnvLogLevel := os.Getenv("LOG_LEVEL"); EnvLogLevel != "" {
		cf.LogLevel = EnvLogLevel
	}

	if EnvFileStorePath := os.Getenv("FILE_STORAGE_PATH"); EnvFileStorePath != "" {
		cf.FileStorePath = EnvFileStorePath
	}

	if EnvDBdsn := os.Getenv("DATABASE_DSN"); EnvDBdsn != "" {
		cf.DBdsn = EnvDBdsn
	}

	if cf.DBdsn != "" {
		cf.StoreType = 1
	}

	return cf
}
