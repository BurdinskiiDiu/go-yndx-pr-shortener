package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/cmd/config"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/logger"
	"go.uber.org/zap"
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
		defer r.Body.Close()
		content, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		longURL := string(content)
		/*var shrtURL string
		cntr := 0
		var errPSU error
		for cntr < 100 {
			shrtURL = shorting()
			if errPSU = uS.PostShortURL(shrtURL, longURL); errPSU != nil {
				cntr++
				continue
			}
			break
		}*/
		shrtURL, errPSU := CreateShortURL(uS, longURL)
		if errPSU != nil {
			log.Println(errPSU.Error())
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
		lngURL, err := uS.GetLongURL(srtURL)
		if err != nil {
			log.Fatal(err.Error())
			return
		}
		w.Header().Set("Location", lngURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
		w.Write([]byte(lngURL))
	})
}

type URLReq struct {
	URL string `json:"url"`
}

type URLResp struct {
	Result string `json:"result"`
}

func PostURLApi(uS URLStore, cf config.Config) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var buf bytes.Buffer
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			logger.Log.Error(err.Error())
		}
		logger.Log.Info("got postApi message", zap.String("body", buf.String()))

		var urlReq URLReq
		if err := json.Unmarshal(buf.Bytes(), &urlReq); err != nil {
			logger.Log.Error(err.Error())
			return
		}
		logger.Log.Info("unmarshaled url from postApi message", zap.String("longURL", urlReq.URL))

		shrtURL, err := CreateShortURL(uS, urlReq.URL)
		if err != nil {
			logger.Log.Error(err.Error())
			return
		}
		logger.Log.Info("short url", zap.String("shortURL", shrtURL))
		var urlResp URLResp
		urlResp.Result = cf.BaseAddr + "/" + shrtURL
		resp, err := json.MarshalIndent(urlResp, "", " ")
		if err != nil {
			logger.Log.Error(err.Error())
			return
		} /*
			respString := string(resp)
			respString = respString[1:]
			respString = respString[:len(respString)-1]
			fmt.Println(string(resp))
			fmt.Println(string(respString))*/
		logger.Log.Info("response for postApi request", zap.String("response", string(resp)))
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", string(resp))
		w.WriteHeader(http.StatusCreated)
		w.Write(resp)
	})
}

func CreateShortURL(uS URLStore, longURL string) (string, error) {
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
	return shrtURL, errPSU
}
