package main

import (
	"log"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/app/handler"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/app/server"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/config"
	filestore "github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/fileStore"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/logg"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/store"
)

func main() {
	conf := config.GetConfig()
	logger, err := logg.InitLog(conf)
	if err != nil {
		log.Fatal(err)
	}
	eF := filestore.CreateFileStore(*conf, logger)
	uS := store.NewURLStorage()
	uS.GetStoreBackup(eF, logger)
	wS := handler.NewWorkStruct(uS, conf)
	rt := server.NewServer(wS, logger)
	rt.Run()
}
