package main

import (
	"context"
	"log"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/app/handler"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/app/server"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/config"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/logg"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/postgresql"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/store"
)

func main() {
	conf := config.GetConfig()
	logger, err := logg.InitLog(conf)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	uS := store.NewURLStorage()
	err = uS.GetStoreBackup(conf, logger)
	if err != nil {
		logger.Fatal(err.Error())
	}

	db := postgresql.NewClientDBStruct(conf.DBdsn, logger)
	err = db.Create()
	if err != nil {
		logger.Fatal(err.Error())
	}

	wS := handler.NewWorkStruct(uS, conf, logger, db, ctx)

	rt := server.NewServer(wS, logger)
	rt.Run()
	defer db.Close()
}
