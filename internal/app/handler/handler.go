package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/config"
	"go.uber.org/zap"
)

type URLStore interface {
	PostShortURL(string, string, *zap.Logger) error
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

func PostLongURL(uS URLStore, cf config.Config, logger *zap.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		content, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		longURL := string(content)
		logger.Info("got post message", zap.String("body", longURL))

		shrtURL, errPSU := CreateShortURL(uS, longURL, logger)
		if errPSU != nil {
			log.Println(errPSU.Error())
		}
		bodyResp := cf.BaseAddr + "/" + shrtURL
		logger.Info("response body message", zap.String("body", bodyResp))
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Content-Length", strconv.Itoa(len(bodyResp)))
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(bodyResp))
	})
}

func GetLongURL(uS URLStore, srtURL string, logger *zap.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("shortURL is:", zap.String("shortURL", srtURL))
		lngURL, err := uS.GetLongURL(srtURL)
		logger.Info("longURL is:", zap.String("longURL", lngURL))
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

func PostURLApi(uS URLStore, cf config.Config, logger *zap.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var buf bytes.Buffer
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			logger.Error(err.Error())
		}
		logger.Info("got postApi message", zap.String("body", buf.String()))

		var urlReq URLReq
		if err := json.Unmarshal(buf.Bytes(), &urlReq); err != nil {
			logger.Error(err.Error())
			return
		}
		logger.Info("unmarshaled url from postApi message", zap.String("longURL", urlReq.URL))

		shrtURL, err := CreateShortURL(uS, urlReq.URL, logger)
		if err != nil {
			logger.Error(err.Error())
			return
		}
		logger.Info("short url", zap.String("shortURL", shrtURL))
		var urlResp URLResp
		urlResp.Result = cf.BaseAddr + "/" + shrtURL

		resp, err := json.Marshal(urlResp)
		logger.Info("resp for postURLApi", zap.String("resp", string(resp)))
		if err != nil {
			logger.Error(err.Error())
			return
		}
		logger.Info("response for postApi request", zap.String("response", string(resp)))
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", strconv.Itoa(len(string(resp))))
		w.WriteHeader(http.StatusCreated)
		w.Write(resp)
	})
}

func CreateShortURL(uS URLStore, longURL string, logger *zap.Logger) (string, error) {
	var shrtURL string
	cntr := 0
	var errPSU error
	for cntr < 100 {
		shrtURL = shorting()
		if errPSU = uS.PostShortURL(shrtURL, longURL, logger); errPSU != nil {
			cntr++
			continue
		}
		break
	}
	return shrtURL, errPSU
}

//handler implementation with methods

type WorkStruct struct {
	US URLStore
	Cf *config.Config
}

func NewWorkStruct(uS URLStore, cf *config.Config) *WorkStruct {
	return &WorkStruct{
		US: uS,
		Cf: cf,
	}
}

func (wS *WorkStruct) PostLongURL(logger *zap.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		content, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		longURL := string(content)
		logger.Info("got post message", zap.String("body", longURL))

		shrtURL, errPSU := CreateShortURL(wS.US, longURL, logger)
		if errPSU != nil {
			log.Println(errPSU.Error())
		}
		bodyResp := wS.Cf.BaseAddr + "/" + shrtURL
		logger.Info("response body message", zap.String("body", bodyResp))
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Content-Length", strconv.Itoa(len(bodyResp)))
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(bodyResp))
	})
}

func (wS *WorkStruct) GetLongURL(srtURL string, logger *zap.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("shortURL is:", zap.String("shortURL", srtURL))
		lngURL, err := wS.US.GetLongURL(srtURL)
		logger.Info("longURL is:", zap.String("longURL", lngURL))
		if err != nil {
			log.Fatal(err.Error())
			return
		}
		w.Header().Set("Location", lngURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
		w.Write([]byte(lngURL))
	})
}

func (wS *WorkStruct) PostURLApi(logger *zap.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var buf bytes.Buffer
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			logger.Error(err.Error())
		}
		logger.Info("got postApi message", zap.String("body", buf.String()))

		var urlReq URLReq
		if err := json.Unmarshal(buf.Bytes(), &urlReq); err != nil {
			logger.Error(err.Error())
			return
		}
		logger.Info("unmarshaled url from postApi message", zap.String("longURL", urlReq.URL))

		shrtURL, err := CreateShortURL(wS.US, urlReq.URL, logger)
		if err != nil {
			logger.Error(err.Error())
			return
		}
		logger.Info("short url", zap.String("shortURL", shrtURL))
		var urlResp URLResp
		urlResp.Result = wS.Cf.BaseAddr + "/" + shrtURL

		resp, err := json.Marshal(urlResp)
		logger.Info("resp for postURLApi", zap.String("resp", string(resp)))
		if err != nil {
			logger.Error(err.Error())
			return
		}
		logger.Info("response for postApi request", zap.String("response", string(resp)))
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", strconv.Itoa(len(string(resp))))
		w.WriteHeader(http.StatusCreated)
		w.Write(resp)
	})
}
