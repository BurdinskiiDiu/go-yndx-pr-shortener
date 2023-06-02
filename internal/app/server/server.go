package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/app/handler"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/config"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Server struct {
	rt chi.Router
	cf *config.Config
}

func NewServer(wS *handler.WorkStruct, logger *zap.Logger) *Server {
	return &Server{
		rt: NewRouter(wS, logger),
		cf: wS.Cf,
	}
}

func ValidConfig(cf *config.Config, logger *zap.Logger) *config.Config {
	da := strings.Split(cf.ServAddr, ":")
	if len(da) == 2 {
		cf.ServAddr = ":" + da[1]
	} else if len(da) == 3 {
		cf.ServAddr = ":" + da[2]
	} else {
		logger.Error("Need address in a form host:port")
		cf.ServAddr = ":8080"
	}

	ba := strings.Split(cf.BaseAddr, ":")
	if len(ba) == 2 {
		cf.BaseAddr = ba[0] + cf.ServAddr
	} else if len(ba) == 3 {
		cf.BaseAddr = ba[0] + ":" + ba[1] + cf.ServAddr
	} else {
		logger.Error("Need address in a form host:port")
		cf.BaseAddr = "http://localhost:8080"
	}
	return cf
}

type ChiData struct {
	shrtURLId string
}

type CompleRespWriter struct {
	http.ResponseWriter
	chiData *ChiData
}

func NewRouter(wS *handler.WorkStruct, logger *zap.Logger) chi.Router {
	wS.Cf = ValidConfig(wS.Cf, logger)
	logger.Info("server starting", zap.String("addr", wS.Cf.ServAddr))
	rt := chi.NewRouter()
	rt.Use(middleware.Timeout(10 * time.Second))
	rt.Use(wS.GZipMiddleware)
	rt.Use(wS.LoggingHandler)
	//rt.Use(wS.GZipMiddleware)
	rt.Post("/", wS.PostLongURL())
	rt.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		logger.Info("chi id is:", zap.String("id", id))
		wS.GetLongURL(id).ServeHTTP(w, r)
	})
	rt.Post("/api/shorten", wS.PostURLApi())
	return rt
}

func (sr *Server) Run() {
	http.ListenAndServe(sr.cf.ServAddr, sr.rt)
}
