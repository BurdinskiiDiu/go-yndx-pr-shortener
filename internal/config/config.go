package config

import (
	"flag"
	"fmt"
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
	flag.StringVar(&cf.DBdsn, "d", "", "dsn dor db connection")
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

	if EnvFileStorePath := os.Getenv("FILE_STORAGE_PATH"); EnvFileStorePath != "" {
		cf.FileStorePath = EnvFileStorePath
	}

	if EnvDBdsn := os.Getenv("DATABASE_DSN"); EnvDBdsn != "" {
		cf.DBdsn = EnvDBdsn
	}

	if cf.DBdsn != "" {
		cf.StoreType = 1
		log.Println("db connString from flag: " + cf.DBdsn)
		/*ps := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		`localhost`, `video`, `XXXXXXXX`, `video`)*/
		dbDsn := "host=" + cf.DBdsn[12:16] + " user=" + cf.DBdsn[3:11] + " dbname=" + cf.DBdsn[17:26] + " sslmode=disable"
		fmt.Println("dbsn from config is: " + dbDsn)
		cf.DBdsn = dbDsn
		log.Println("new db connString from flag: " + cf.DBdsn)
	}

	return cf
}
