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

	mapStore := store.NewURLStorage(logger)
	var store handler.URLStore
	if conf.StoreType != 0 {
		dbStore := postgresql.NewClientDBStruct(ctx, logger, conf)
		err = dbStore.Create()
		defer dbStore.Close()
		if err != nil {
			logger.Error("creating db err", zap.Error(err))
			conf.StoreType = 0
			store = mapStore
		} else {
			store = dbStore
		}

	} else {
		store = mapStore
	}

	wS := handler.NewWorkStruct(store, conf, logger, ctx)
	err = wS.GetStoreBackup()
	if err != nil {
		logger.Fatal(err.Error())
	}
	rt := server.NewServer(wS, logger)
	rt.Run()

}
