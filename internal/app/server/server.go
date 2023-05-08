package server

import (
	"net/http"
	"time"
)

type Server struct {
	srv http.Server
}

func NewServer(addr string, h http.Handler) *Server {
	s := &Server{}

	s.srv = http.Server{
		Addr:         addr,
		Handler:      h,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	return s
}

func (s *Server) Run() {
	s.srv.ListenAndServe()
}

/*func Run(adr string, rt http.ServeMux) error {
	return http.ListenAndServe(adr, mux)
}*/
