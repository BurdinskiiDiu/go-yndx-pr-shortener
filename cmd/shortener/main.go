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
	"go.uber.org/zap"
)

func main() {
	conf := config.GetConfig()
	logger, err := logg.InitLog(conf)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	uS := store.NewURLStorage(logger)

	db := postgresql.NewClientDBStruct(ctx, logger, conf)
	err = db.Create()
	if err != nil {
		logger.Error("creating db err", zap.Error(err))
		conf.StoreType = 0
	}

	wS := handler.NewWorkStruct(uS, conf, logger, db, ctx)
	err = wS.GetStoreBackup()
	if err != nil {
		logger.Fatal(err.Error())
	}
	rt := server.NewServer(wS, logger)
	rt.Run()
	defer db.Close()
}
