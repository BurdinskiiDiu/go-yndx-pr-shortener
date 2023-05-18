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

func GetConfig() Config {
	cf := new(Config)
	flag.StringVar(&cf.ServAddr, "a", ":8080", "default HTTP-server addres")
	flag.StringVar(&cf.BaseAddr, "b", "http://localhost:8080", "base host addr for short URL response")
	flag.Parse()

	if cf.EnvServAddr = os.Getenv("SERVER_ADDRESS"); cf.EnvServAddr != "" {
		cf.ServAddr = cf.EnvServAddr
	}

	if cf.EnvBaseAddr = os.Getenv("BASE_URL"); cf.EnvBaseAddr != "" {
		cf.BaseAddr = cf.EnvBaseAddr
	}
	return *cf
	/*
		da := strings.Split(cf.DefaultAddr, ":")
		if len(da) == 2 {
			cf.DefaultAddr = ":" + da[1]
		} else if len(da) == 3 {
			cf.DefaultAddr = ":" + da[2]
		} else {
			log.Printf("Need address in a form host:port")
			cf.DefaultAddr = ":8080"
		}

		ba := strings.Split(cf.BaseAddr, ":")
		if len(ba) == 2 {
			cf.BaseAddr = ba[0] + cf.DefaultAddr
		} else if len(ba) == 3 {
			cf.BaseAddr = ba[0] + ":" + ba[1] + cf.DefaultAddr
		} else {
			log.Printf("Need address in a form host:port")
			cf.BaseAddr = "http://localhost:8080"
		}*/
}
