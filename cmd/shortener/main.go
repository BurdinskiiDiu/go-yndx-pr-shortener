package main

import (
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/cmd/store"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/app/handler"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/app/server"
)

func main() {
	uS := store.NewUrlStorage()
	h := handler.NewRouter(uS)
	srv := server.NewServer(":8080", h)
	srv.Run()
}
