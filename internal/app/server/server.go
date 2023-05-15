package server

import (
	"net/http"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/cmd/store"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/app/handler"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	rt   chi.Router
	addr string
}

func NewServer(uS store.URLStore, addr string) *Server {
	return &Server{
		rt:   NewRouter(uS),
		addr: addr,
	}
}

func NewRouter(uS store.URLStore) chi.Router {
	rt := chi.NewRouter()

	rt.Post("/", handler.PostLongURL(uS))
	rt.Get("/{id}", handler.GetLongURL(uS))
	return rt
}
func (sr *Server) Run() {
	http.ListenAndServe(sr.addr, sr.rt)
}
