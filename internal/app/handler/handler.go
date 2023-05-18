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
	AddShortURL(string, string) bool
	GetLongURL(string) (string, error)
}

/*
type Router struct {
	*http.ServeMux
	uS URLStore
}*/

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

		//defer r.Body.Close()
		content, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		longURL := string(content)
		shrtURL := ""
		done := false
		iterCheck := 0
		for !done {
			shrtURL = shorting()
			done = uS.AddShortURL(shrtURL, longURL)
			iterCheck++
			if iterCheck > 1000 {
				log.Fatal("shortURLstore is full, please expand store")
			}
		}
		/*
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}*/
		bodyResp := cf.BaseAddr + "/" + shrtURL
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Content-Length", strconv.Itoa(len(bodyResp+" long url is "+longURL)))
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(bodyResp + " long url is " + longURL))

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
			http.Error(w, err.Error()+" "+lngURL, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Location", lngURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
		w.Write([]byte("Location: " + lngURL))

	})
}

/*
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
}*/
