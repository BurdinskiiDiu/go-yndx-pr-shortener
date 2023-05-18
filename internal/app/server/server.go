package server

import (
	"net/http"
	"time"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/cmd/config"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/cmd/store"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/app/handler"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	rt   chi.Router
	conf config.Config
}

func NewServer(uS store.URLStore, conf config.Config) *Server {
	return &Server{
		rt:   NewRouter(uS, conf),
		conf: conf,
	}
}

func NewRouter(uS store.URLStore, conf config.Config) chi.Router {
	rt := chi.NewRouter()
	rt.Use(middleware.Timeout(10 * time.Second))
	rt.Post("/", handler.PostLongURL(uS, conf))
	rt.Get("/{id}", handler.GetLongURL(uS))
	return rt
}
func (sr *Server) Run() {
	http.ListenAndServe(sr.conf.DefaultAddr, sr.rt)
}
