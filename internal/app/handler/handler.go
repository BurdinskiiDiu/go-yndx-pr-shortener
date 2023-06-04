package handler

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/config"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/gzp"
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

type URLReq struct {
	URL string `json:"url"`
}

type URLResp struct {
	Result string `json:"result"`
}

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
		wS.logger.Info("response body message", zap.String("body", bodyResp))
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
		wS.logger.Debug("longURL is:", zap.String("longURL", lngURL))
		w.Write([]byte(lngURL))
	})
}

func (wS *WorkStruct) PostURLApi() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		accptEnc := r.Header.Get("Accept-Encoding")
		wS.logger.Info("acceptEnc", zap.String("accptEnc", accptEnc))
		contType := r.Header.Get("Content-Type")
		wS.logger.Info("contType", zap.String("contType", contType))
		//permissContType := strings.Contains(contType, "text/plain") || strings.Contains(contType, "application/json")
		suppGZip := strings.Contains(accptEnc, "gzip")
		if suppGZip /*&& permissContType*/ {
			//cw := gzp.NewCompressWriter(w)
			//w.Header.("Content-Encoding", "gzip")
			gz := gzip.NewWriter(w)
			cw := gzp.NewCompressWriter(w, gz)
			ow = cw
			ow.Header().Set("Content-Encoding", "gzip")
			defer cw.Close()
		}
		wS.logger.Info("acceptEnc", zap.String("accptEnc", accptEnc))
		cntntEnc := r.Header.Get("Content-Encoding")
		wS.logger.Info("cntntEnc", zap.String("cntntEnc", cntntEnc))
		sendGZip := strings.Contains(cntntEnc, "gzip")
		if sendGZip {
			cr, err := gzp.NewCompressReader(r.Body)
			if err != nil {
				wS.logger.Error("compersReader creation err", zap.Error(err))
				return
			}
			r.Body = cr
			defer cr.Close()
		}
		wS.logger.Info("response", zap.String("response", r.RequestURI))

		h.ServeHTTP(ow, r)
	})
}
