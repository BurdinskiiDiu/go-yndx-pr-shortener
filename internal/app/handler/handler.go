package handler

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/cmd/config"
)

/*
type Router struct {
	*http.ServeMux
	uS URLStore
}*/

type URLStore interface {
	PostShortURL(string, string) bool
	GetLongURL(string) (string, error)
}

const letterBytes = "abcdifghijklmnopqrstuvwxyzABCDIFGHIJKLMNOPQRSTUVWXYZ"

func shorting() string {
	b := make([]byte, 8)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func PostLongURL(uS URLStore, cf config.Config) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "bad method", http.StatusBadRequest)
			return
		}

		defer r.Body.Close()
		content, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		longURL := string(content)
		var shrtURL string

		//shrtURL, err := uS.CreateShortURL(string(content))
		done := false
		for !done {
			shrtURL = shorting()
			done = uS.PostShortURL(shrtURL, longURL)
		}
		/*
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}*/
		bodyResp := cf.BaseAddr + "/" + shrtURL
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Content-Length", strconv.Itoa(len( /*cf.BaseAddr+"/"+shrtURL*/ bodyResp)))
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte( /*cf.BaseAddr + "/" + shrtURL*/ bodyResp))
	})
}

func GetLongURL(uS URLStore) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "bad method", http.StatusBadRequest)
			return
		}
		srtURL := r.URL.Path
		srtURL = srtURL[1:]
		lngURL, err := uS.GetLongURL(srtURL)
		if err != nil {
			log.Fatal(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Location", lngURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
		//w.Write([]byte("Location: " + lngURL))
	})
}
