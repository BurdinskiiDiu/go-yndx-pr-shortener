package server

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/cmd/config"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/app/handler"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	rt   chi.Router
	conf config.Config
}

func CheckConfig(cf *config.Config) {
	servAddrSlice := strings.Split(cf.ServAddr, ":")
	if len(servAddrSlice) == 2 {
		cf.ServAddr = ":" + servAddrSlice[1]
	} else if len(servAddrSlice) == 3 {
		cf.ServAddr = ":" + servAddrSlice[2]
	} else {
		log.Printf("Need address in a form host:port")
		cf.ServAddr = ":8080"
	}

	baseAddrSlice := strings.Split(cf.BaseAddr, ":")
	if len(baseAddrSlice) == 2 {
		cf.BaseAddr = baseAddrSlice[0] + cf.ServAddr
	} else if len(baseAddrSlice) == 3 {
		cf.BaseAddr = baseAddrSlice[0] + ":" + baseAddrSlice[1] + cf.ServAddr
	} else {
		log.Printf("Need address in a form host:port")
		cf.BaseAddr = "http://localhost:8080"
	}
}

func NewServer(uS handler.URLStore, conf config.Config) *Server {
	return &Server{
		rt:   NewRouter(uS, conf),
		conf: conf,
	}
}

/*
func GetUrlFromChi(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}*/

func NewRouter(uS handler.URLStore, conf config.Config) *chi.Mux {
	rt := chi.NewRouter()
	rt.Use(middleware.Timeout(10 * time.Second))

	rt.Post("/", handler.PostLongURL(uS, conf))
	rt.Get("/{id}", handler.GetLongURL(uS))
	return rt
}

func (sr *Server) Run() {
	http.ListenAndServe(sr.conf.ServAddr, sr.rt)
}
