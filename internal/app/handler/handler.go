package handler

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/cmd/config"
)

type URLStore interface {
	PostShortURL(string, string) error
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
		cntr := 0
		var errPSU error
		for cntr < 100 {
			shrtURL = shorting()
			if errPSU = uS.PostShortURL(shrtURL, longURL); errPSU != nil {
				cntr++
				continue
			}
			break
		}
		if errPSU != nil {
			http.Error(w, errPSU.Error(), http.StatusInternalServerError)
		}
		bodyResp := cf.BaseAddr + "/" + shrtURL
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Content-Length", strconv.Itoa(len(bodyResp)))
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(bodyResp))
	})
}

func GetLongURL(uS URLStore, srtURL string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "bad method", http.StatusBadRequest)
			return
		}
		//srtURL := r.URL.Path
		//srtURL := chi.URLParam(r, "id")
		//srtURL = srtURL[1:]
		lngURL, err := uS.GetLongURL(srtURL)
		if err != nil {
			log.Fatal(err.Error())
			return
		}
		w.Header().Set("Location", lngURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	})
}
