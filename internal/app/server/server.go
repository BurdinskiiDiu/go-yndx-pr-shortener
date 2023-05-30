package server

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/cmd/config"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/cmd/gzip"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/app/handler"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/logger"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Server struct {
	rt   chi.Router
	conf config.Config
}

func NewServer(uS handler.URLStore, conf config.Config) *Server {
	return &Server{
		rt:   NewRouter(uS, conf),
		conf: conf,
	}
}

func ValidConfig(cf *config.Config) config.Config {
	da := strings.Split(cf.ServAddr, ":")
	if len(da) == 2 {
		cf.ServAddr = ":" + da[1]
	} else if len(da) == 3 {
		cf.ServAddr = ":" + da[2]
	} else {
		log.Printf("Need address in a form host:port")
		cf.ServAddr = ":8080"
	}

	ba := strings.Split(cf.BaseAddr, ":")
	if len(ba) == 2 {
		cf.BaseAddr = ba[0] + cf.ServAddr
	} else if len(ba) == 3 {
		cf.BaseAddr = ba[0] + ":" + ba[1] + cf.ServAddr
	} else {
		log.Printf("Need address in a form host:port")
		cf.BaseAddr = "http://localhost:8080"
	}
	return *cf
}

type ChiData struct {
	shrtURLId string
}

type CompleRespWriter struct {
	http.ResponseWriter
	chiData *ChiData
}

func NewRouter(uS handler.URLStore, conf config.Config) chi.Router {
	conf = ValidConfig(&conf)
	logger.Log.Info("server starting", zap.String("addr", conf.ServAddr))
	rt := chi.NewRouter()
	rt.Use(middleware.Timeout(10 * time.Second))
	rt.Post("/", logger.LoggingHandler(gzip.GZipMiddleware(handler.PostLongURL(uS, conf).ServeHTTP)))
	rt.Get("/{id}", logger.LoggingHandler(gzip.GZipMiddleware(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		logger.Log.Info("chi id is:", zap.String("id", id))
		handler.GetLongURL(uS, id).ServeHTTP(w, r)
	})))
	rt.Post("/api/shorten", logger.LoggingHandler(gzip.GZipMiddleware(handler.PostURLApi(uS, conf).ServeHTTP)))
	return rt
}

func (sr *Server) Run() {
	http.ListenAndServe(sr.conf.ServAddr, sr.rt)
}
