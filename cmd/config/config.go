package config

import (
	"flag"
	"os"
)

type Config struct {
	ServAddr    string
	BaseAddr    string
	EnvServAddr string
	EnvBaseAddr string
}

func GetConfig() *Config {
	cf := new(Config)
	flag.StringVar(&cf.ServAddr, "a", ":8080", "default HTTP-server addres")
	flag.StringVar(&cf.BaseAddr, "b", "http://localhost:8080", "base host addr for short URL response")
	flag.Parse()

	if EnvServAddr := os.Getenv("SERVER_ADDRESS"); EnvServAddr != "" {
		cf.ServAddr = EnvServAddr
	}

	if EnvBaseAddr := os.Getenv("BASE_URL"); EnvBaseAddr != "" {
		cf.BaseAddr = EnvBaseAddr
	}
	/*
		da := strings.Split(cf.ServAddr, ":")
		if len(da) == 2 {
			cf.ServAddr = ":" + da[1]
		} else if len(da) == 3 {
			cf.ServAddr = ":" + da[2]
		} else {
			fmt.Println("Need address in a form host:port")
			cf.ServAddr = ":8080"
		}

		ba := strings.Split(cf.BaseAddr, ":")
		if len(ba) == 2 {
			cf.BaseAddr = ba[0] + cf.ServAddr
		} else if len(ba) == 3 {
			cf.BaseAddr = ba[0] + ":" + ba[1] + cf.ServAddr
		} else {
			fmt.Println("Need address in a form host:port")
			cf.BaseAddr = "http://localhost:8080"
		}*/
	/*
		fmt.Println(cf.ServAddr)
		fmt.Println(cf.BaseAddr)*/
	return cf
}
