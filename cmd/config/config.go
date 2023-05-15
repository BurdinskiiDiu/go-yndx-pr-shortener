package config

import (
	"flag"
	"fmt"
	"net/url"
)

type Config struct {
	DefaultAddr string
	BaseAddr    string
}

func ParseFlags(cf *Config) {
	flag.StringVar(&cf.DefaultAddr, "a", ":8080", "default HTTP-server addres")
	flag.StringVar(&cf.BaseAddr, "b", "http://localhost:8080", "base host addr for short URL response")
	flag.Parse()
	fmt.Println(cf.DefaultAddr)
	fmt.Println(cf.BaseAddr)
	//isUrl1, _ := isUrl(cf.DefaultAddr)
	//isUrl2, _ := isUrl(cf.BaseAddr)
	/*
		if !isUrl1 {
			cf.DefaultAddr = ":8080"
		}

		if !isUrl2 {
			cf.BaseAddr = "http://localhost:8080"
		}*/
	//fmt.Println(cf.DefaultAddr)
	//fmt.Println(cf.BaseAddr)

}

func isUrl(str string) (bool, error) {
	parsedURL, err := url.Parse(str)
	return err == nil && parsedURL.Scheme != "" && parsedURL.Host != "", err
	/*url, err := url.ParseRequestURI(str)
	if err != nil {
		fmt.Println(err)
	}
	if url.Scheme == "" && str[:2] != "://" {
		str = "http://" + str
	}

	if url.Host == "" {
		url.Host = "localhost:8080"
	}
	return url, nil*/
}
