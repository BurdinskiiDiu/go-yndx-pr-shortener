package main

import (
	"log"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/cmd/config"
	filestore "github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/cmd/fileStore"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/cmd/store"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/app/handler"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/app/server"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/logger"
)

func main() {
	conf := config.GetConfig()
	err := logger.InitLog(conf)
	if err != nil {
		log.Fatal(err)
	}
	eF := filestore.CreateFileStore(*conf)
	uS := store.NewURLStorage()
	uS.GetStoreBackup(eF)
	wS := handler.NewWorkStruct(uS, conf)
	rt := server.NewServer(wS)
	rt.Run()
}
