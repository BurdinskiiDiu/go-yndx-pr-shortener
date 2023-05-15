package config

import (
	"flag"
	"fmt"
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
	hp := strings.Split(cf.DefaultAddr, ":")
	if len(hp) != 2 {
		fmt.Println("Need address in a form host:port")
		return
	}
	cf.DefaultAddr = ":" + hp[1]
	fmt.Println(cf.DefaultAddr)
	fmt.Println(cf.BaseAddr)
}
