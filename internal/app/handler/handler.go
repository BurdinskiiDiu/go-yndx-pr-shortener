package handler

import (
	"io"
	"net/http"
	"strconv"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/cmd/config"
)

type Router struct {
	*http.ServeMux
	uS URLStore
}

type URLStore interface {
	CreateShortURL(string) (string, error)
	GetLongURL(string) (string, error)
}

func PostLongURL(uS URLStore, cf config.Config) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			defer r.Body.Close()

			content, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			shrtURL, err := uS.CreateShortURL(string(content))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "text/plain; charset=utf-8")

			w.Header().Set("Content-Length", strconv.Itoa(len(cf.BaseAddr+"/"+shrtURL)))
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(cf.BaseAddr + "/" + shrtURL))
		} else {
			http.Error(w, "bad method", http.StatusBadRequest)
			return
		}
	})
}

func GetLongURL(uS URLStore) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			srtURL := r.URL.Path
			srtURL = srtURL[1:]
			lngURL, err := uS.GetLongURL(srtURL)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Location", lngURL)
			w.WriteHeader(http.StatusTemporaryRedirect)
			//w.Write([]byte("Location: " + lngURL))
		} else {
			http.Error(w, "bad method", http.StatusBadRequest)
			return
		}
	})
}
