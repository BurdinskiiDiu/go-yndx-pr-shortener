package main

import (
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/cmd/store"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/app/server"
)

func main() {
	uS := store.NewURLStorage()
	//h := handler.NewRouter(uS)
	rt := server.NewServer(uS, ":8080")
	//srv := server.NewServer(":8080", h)
	rt.Run()
}
