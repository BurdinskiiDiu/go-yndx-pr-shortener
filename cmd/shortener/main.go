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

	//mapStore := store.NewURLStorage(logger)
	var str handler.URLStore
	if conf.StoreType != 0 {
		dbStore := postgresql.NewClientDBStruct( /*ctx,*/ logger, conf)
		//defer dbStore.Close(ctx)
		err = dbStore.Create(ctx)

		if err != nil {
			//logger.Error("creating db err", zap.Error(err))
			//conf.StoreType = 0
			//str = mapStore
			logger.Fatal("creating db err", zap.Error(err))
			//} else {

		}
		str = dbStore

	} else {
		mapStore := store.NewURLStorage(logger)
		str = mapStore
	}

	hn := handler.NewHandlers( /*ctx,*/ str, conf, logger)
	if conf.StoreType == 0 {
		err = hn.GetStoreBackup()
		if err != nil {
			logger.Fatal(err.Error())
			logger.Error(err.Error())
		}
	}

	rt := server.NewServer(hn, logger)
	rt.Run()

}
