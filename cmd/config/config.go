package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	DefaultAddr string
	BaseAddr    string
}

func ParseFlags(cf *Config) {
	flag.StringVar(&cf.DefaultAddr, "a", ":8080", "default HTTP-server addres")

	flag.StringVar(&cf.BaseAddr, "b", "http://localhost:8080", "base host addr for short URL response")
	flag.Parse()

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		cf.DefaultAddr = envRunAddr
	}

	if envBaseAddr := os.Getenv("BASE_URL"); envBaseAddr != "" {
		cf.BaseAddr = envBaseAddr
	}

	da := strings.Split(cf.DefaultAddr, ":")
	if len(da) == 2 {
		cf.DefaultAddr = ":" + da[1]
	} else if len(da) == 3 {
		cf.DefaultAddr = ":" + da[2]
	} else {
		fmt.Println("Need address in a form host:port")
		cf.DefaultAddr = ":8080"
	}

	ba := strings.Split(cf.BaseAddr, ":")
	if len(ba) == 2 {
		cf.BaseAddr = ba[0] + cf.DefaultAddr
	} else if len(ba) == 3 {
		cf.BaseAddr = ba[0] + ":" + ba[1] + cf.DefaultAddr
	} else {
		fmt.Println("Need address in a form host:port")
		cf.BaseAddr = "http://localhost:8080"
	}

	fmt.Println(cf.DefaultAddr)
	fmt.Println(cf.BaseAddr)
}
