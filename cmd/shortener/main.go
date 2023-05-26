package main

import (
	"log"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/cmd/config"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/cmd/store"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/app/handler"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/app/server"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/logger"
)

func main() {
	conf := config.GetConfig()
	err := logger.InitLog(conf)
	if err != nil {
		log.Fatal(err.Error())
	}

	uS := store.NewURLStorage()
	wS := handler.NewWorkStruct(uS, conf)
	rt := server.NewServer(wS)
	rt.Run()
}
