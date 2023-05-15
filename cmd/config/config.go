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
	//cf.DefaultAddr = "http://" + cf.DefaultAddr
	hp := strings.Split(cf.DefaultAddr, ":")
	if len(hp) == 2 {
		cf.DefaultAddr = ":" + hp[1]

	} else if len(hp) == 3 {
		cf.DefaultAddr = ":" + hp[2]

	} else {

		fmt.Println("Need address in a form host:port")
		cf.DefaultAddr = ":8080"
		return
	}

	fmt.Println(cf.DefaultAddr)
	fmt.Println(cf.BaseAddr)
}
