package main

import (
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/cmd/config"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/cmd/store"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/app/server"
)

func main() {
	conf := &config.Config{}
	config.ParseFlags(conf)
	uS := store.NewURLStorage()
	//h := handler.NewRouter(uS)
	rt := server.NewServer(uS, *conf)
	//srv := server.NewServer(":8080", h)
	rt.Run()
}
