package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/config"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/handler"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/logg"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/postgresql"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/server"
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
	var str handler.URLStore
	if conf.StoreType != 0 {
		tm := 1 * time.Minute
		dbStore := postgresql.NewClientDBStruct(logger, conf, tm)
		defer dbStore.Close()
		err = dbStore.Create(ctx)

		if err != nil {
			logger.Fatal("creating db err", zap.Error(err))
		}
		str = dbStore

	} else {
		mapStore := store.NewURLStorage(logger)
		str = mapStore
	}
	delURLSChan := make(chan []postgresql.URLsForDel, conf.DelChnlLen)
	defer close(delURLSChan)
	delMtx := new(sync.Mutex)
	hn := handler.NewHandlers(str, delURLSChan, conf, logger, delMtx)
	if conf.StoreType == 0 {
		err = hn.GetStoreBackup()
		if err != nil {
			logger.Fatal(err.Error())
			logger.Error(err.Error())
		}
	}
	rt := server.NewServer(hn, logger)
	go hn.DelURLSBatch()
	rt.Run()

}
