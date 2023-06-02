package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/config"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/gzip"
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

/*
func PostLongURL(uS URLStore, cf config.Config, logger *zap.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//defer r.Body.Close()
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
}*/

type URLReq struct {
	URL string `json:"url"`
}

type URLResp struct {
	Result string `json:"result"`
}

/*
func PostURLApi(uS URLStore, cf config.Config, logger *zap.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//defer r.Body.Close()
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
}*/
/*func CreateShortURL(uS URLStore, longURL string, logger *zap.Logger) (string, error) {
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
}*/
//handler implementation with methods

type WorkStruct struct {
	US     URLStore
	Cf     *config.Config
	logger *zap.Logger
}

func (wS *WorkStruct) CreateShortURL(longURL string) (string, error) {
	var shrtURL string
	cntr := 0
	var errPSU error
	for cntr < 100 {
		shrtURL = shorting()
		if errPSU = wS.US.PostShortURL(shrtURL, longURL, wS.logger); errPSU != nil {
			cntr++
			continue
		}
		break
	}
	return shrtURL, errPSU
}

func NewWorkStruct(uS URLStore, cf *config.Config, logger *zap.Logger) *WorkStruct {
	return &WorkStruct{
		US:     uS,
		Cf:     cf,
		logger: logger,
	}
}

func (wS *WorkStruct) PostLongURL() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//defer r.Body.Close()
		content, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		longURL := string(content)
		wS.logger.Debug("got post message", zap.String("body", longURL))

		shrtURL, err := wS.CreateShortURL(longURL)
		if err != nil {
			wS.logger.Error("error while crearing shortURL", zap.Error(err))
		}
		bodyResp := wS.Cf.BaseAddr + "/" + shrtURL
		wS.logger.Debug("response body message", zap.String("body", bodyResp))
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Content-Length", strconv.Itoa(len(bodyResp)))
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(bodyResp))
	})
}

func (wS *WorkStruct) GetLongURL(srtURL string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wS.logger.Debug("shortURL is:", zap.String("shortURL", srtURL))
		lngURL, err := wS.US.GetLongURL(srtURL)
		wS.logger.Debug("longURL is:", zap.String("longURL", lngURL))
		if err != nil {
			wS.logger.Error("getLongURL handler, error while getting long url from store", zap.Error(err))
			return
		}
		w.Header().Set("Location", lngURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
		w.Write([]byte(lngURL))
	})
}

func (wS *WorkStruct) PostURLApi() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//defer r.Body.Close()
		var buf bytes.Buffer
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			wS.logger.Error("posrURLApi handler, read from request body err", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		wS.logger.Debug("got postApi message", zap.String("body", buf.String()))

		var urlReq URLReq
		if err := json.Unmarshal(buf.Bytes(), &urlReq); err != nil {
			wS.logger.Error("postURLApi handler, unmarshal func err", zap.Error(err))
			return
		}
		wS.logger.Debug("unmarshaled url from postApi message", zap.String("longURL", urlReq.URL))

		shrtURL, err := wS.CreateShortURL(urlReq.URL)
		if err != nil {
			wS.logger.Error("postURLApi handler, creating short url err", zap.Error(err))
			return
		}
		wS.logger.Debug("short url", zap.String("shortURL", shrtURL))
		var urlResp URLResp
		urlResp.Result = wS.Cf.BaseAddr + "/" + shrtURL
		resp, err := json.Marshal(urlResp)
		if err != nil {
			wS.logger.Error("postURLApi handler, marshal func error", zap.Error(err))
			return
		}

		wS.logger.Debug("response for postApi request", zap.String("response", string(resp)))
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", strconv.Itoa(len(string(resp))))
		w.WriteHeader(http.StatusCreated)
		w.Write(resp)
	})
}

//log middleware

type (
	responseData struct {
		status int
		size   int
	}
	LoggingRespWrt struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (lRW *LoggingRespWrt) Write(b []byte) (int, error) {
	size, err := lRW.ResponseWriter.Write(b)
	if err != nil {
		return 0, fmt.Errorf("logger internal err: %w", err)
	}
	lRW.responseData.size += size
	return size, nil
}

func (lRW *LoggingRespWrt) WriteHeader(stCode int) {
	lRW.ResponseWriter.WriteHeader(stCode)
	lRW.responseData.status = stCode
}

func (wS *WorkStruct) LoggingHandler(h http.Handler) http.Handler {
	//func LoggingHandler(h http.Handler, logger *zap.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		responseData := &responseData{
			status: 0,
			size:   0,
		}

		lgRspWrt := LoggingRespWrt{
			ResponseWriter: w,
			responseData:   responseData,
		}
		start := time.Now()
		h.ServeHTTP(&lgRspWrt, r)
		duration := time.Since(start)

		wS.logger.Info("incoming request data",
			zap.String("URl", r.RequestURI),
			zap.String("method", r.Method),
			zap.Int("status", responseData.status),
			zap.Int("size", responseData.size),
			zap.Int("duration", int(duration.Milliseconds())),
		)
	})
}

//log gzip

func (wS *WorkStruct) GZipMiddleware(h http.Handler) http.Handler {
	//func GZipMiddleware(h http.HandlerFunc, logger *zap.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		accptEnc := r.Header.Get("Accept-Encoding")
		wS.logger.Info("acceptEnc", zap.String("accptEnc", accptEnc))
		contType := r.Header.Get("Content-Type")
		wS.logger.Info("contType", zap.String("contType", contType))
		permissContType := strings.Contains(contType, "text/plain") || strings.Contains(contType, "application/json")
		suppGZip := strings.Contains(accptEnc, "gzip")
		if suppGZip && permissContType {
			cw := gzip.NewCompressWriter(ow)
			ow = cw
			defer cw.Close()
		}
		wS.logger.Debug("acceptEnc", zap.String("accptEnc", accptEnc))
		cntntEnc := r.Header.Get("Content-Encoding")
		wS.logger.Debug("cntntEnc", zap.String("cntntEnc", cntntEnc))
		sendGZip := strings.Contains(cntntEnc, "gzip")
		if sendGZip {
			cr, err := gzip.NewCompressReader(r.Body)
			if err != nil {
				wS.logger.Error("compersReader creation err", zap.Error(err))
				return
			}
			r.Body = cr
			defer cr.Close()
		}
		wS.logger.Debug("response", zap.String("response", r.RequestURI))

		h.ServeHTTP(ow, r)
	})
}
