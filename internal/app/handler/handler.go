package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/config"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/gzp"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/postgresql"
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
	db     *postgresql.ClientDBStruct
	ctx    context.Context
}

func NewWorkStruct(uS URLStore, cf *config.Config, logger *zap.Logger, db *postgresql.ClientDBStruct, ctx context.Context) *WorkStruct {
	return &WorkStruct{
		US:     uS,
		Cf:     cf,
		logger: logger,
		db:     db,
		ctx:    ctx,
	}
}

func (wS *WorkStruct) CreateShortURL(longURL string) (string, error) {
	var shrtURL string
	cntr := 0
	var errPSU error
	//existing := errors.New("this short url is already involved")
	//shrtURL = shorting()
	var fn func(string, string) error
	switch wS.Cf.StoreType {
	case 1:
		fn = wS.db.PostShortURL
	default:
		fn = wS.US.PostShortURL
	}

	for cntr < 100 {
		shrtURL = shorting()
		if errPSU = fn(shrtURL, longURL); errPSU != nil {
			cntr++
			continue
			/*if errPSU == existing {

			}*/
			//return "", errPSU
		}
		break
	}
	return shrtURL, errPSU
}

func (wS *WorkStruct) PostLongURL() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wS.logger.Info("start post request")
		content, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		longURL := string(content)
		wS.logger.Info("got post message", zap.String("body", longURL))
		var shrtURL string
		shrtURL, err = wS.CreateShortURL(longURL)
		if err != nil {
			wS.logger.Error("error while crearing shortURL", zap.Error(err))
		}
		bodyResp := wS.Cf.BaseAddr + "/" + shrtURL
		wS.logger.Info("response body message", zap.String("body", bodyResp))
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(bodyResp))
	})
}

func (wS *WorkStruct) GetLongURL(srtURL string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wS.logger.Debug("shortURL is:", zap.String("shortURL", srtURL))

		var fn func(string) (string, error)
		switch wS.Cf.StoreType {
		case 1:
			fn = wS.db.GetLongURL
		default:
			fn = wS.US.GetLongURL
		}

		lngURL, err := fn(srtURL)
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
		wS.logger.Debug("acceptEnc", zap.String("accptEnc", accptEnc))
		contType := r.Header.Get("Content-Type")
		wS.logger.Debug("contType", zap.String("contType", contType))
		suppGZip := strings.Contains(accptEnc, "gzip")
		if suppGZip {
			cw := gzp.NewCompressWriter(w)
			ow = cw
		}
		wS.logger.Debug("acceptEnc", zap.String("accptEnc", accptEnc))
		cntntEnc := r.Header.Get("Content-Encoding")
		wS.logger.Debug("cntntEnc", zap.String("cntntEnc", cntntEnc))
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
		wS.logger.Debug("response", zap.String("response", r.RequestURI))

		h.ServeHTTP(ow, r)
	})
}

//db handler

func (wS *WorkStruct) GetDBPing() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := wS.db.Ping(); err != nil {
			wS.logger.Error("getDBping handler error", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}
