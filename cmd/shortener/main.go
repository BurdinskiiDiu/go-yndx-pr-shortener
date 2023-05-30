package main

import (
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/cmd/config"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/cmd/store"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/app/server"
)

func main() {
	conf := config.GetConfig()
	uS := store.NewURLStorage()
	rt := server.NewServer(uS, *conf)
	rt.Run()
}
