package handler

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/config"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/gzp"
	"go.uber.org/zap"
)

type URLStore interface {
	PostShortURL(string, string, int32) error
	GetLongURL(string) (string, error)
	Ping() error
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
	ctx    context.Context
	uuid   int32
}

func NewWorkStruct(uS URLStore, cf *config.Config, logger *zap.Logger, ctx context.Context) *WorkStruct {
	return &WorkStruct{
		US:     uS,
		Cf:     cf,
		logger: logger,
		ctx:    ctx,
		uuid:   1,
	}
}

func (wS *WorkStruct) CreateShortURL(longURL string) (string, error) {
	var shrtURL string
	cntr := 0
	var errPSU error
	wS.uuid++
	for cntr < 100 {
		shrtURL = shorting()
		if errPSU = wS.US.PostShortURL(shrtURL, longURL, wS.uuid); errPSU != nil {
			cntr++
			continue
		}
		break
	}
	if err := wS.FileFilling(shrtURL, longURL); err != nil {
		wS.logger.Error("createShortURL method, err while filling file", zap.Error(err))
		return "", err
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
		wS.logger.Info("got post message" + longURL)
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
		if err := wS.US.Ping(); err != nil {
			wS.logger.Error("getDBping handler error", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}

//workStruct init

type URLDataStruct struct {
	UUID    string `json:"uuid"`
	ShrtURL string `json:"short_url"`
	LngURL  string `json:"original_url"`
}

func (wS *WorkStruct) GetStoreBackup() error {

	wS.logger.Debug("storefile addr from createfile", zap.String("path", wS.Cf.FileStorePath))

	file, err := os.OpenFile(wS.Cf.FileStorePath, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		wS.logger.Error("open storeFile error")
		return fmt.Errorf("open store_file error: %w", err)

	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	urlDataStr := new(URLDataStruct)
	var raw string
	for scanner.Scan() {
		raw = scanner.Text()
		err := json.Unmarshal([]byte(raw), urlDataStr)
		if err != nil {
			wS.logger.Error("unmarhalling store_file error")
			return err
		}
		uuid, err := strconv.Atoi(urlDataStr.UUID)
		if err != nil {
			return errors.New("error while filling db from backup file, uuid conv to int err:" + err.Error())
		}
		err = wS.US.PostShortURL(urlDataStr.ShrtURL, urlDataStr.LngURL, int32(uuid))
		if err != nil {
			wS.logger.Error("getStoreBackup error, try to write itno db", zap.Error(err))
		}
	}

	if urlDataStr.UUID != "" {
		uuid, err := strconv.Atoi(urlDataStr.UUID)
		if err != nil {
			wS.logger.Error("gettitng last uuid error, file is damaged")
			return err
		}
		wS.uuid = int32(uuid)
	}
	return nil
}

func (wS *WorkStruct) FileFilling(shrtURL, lngURL string) error {
	wS.logger.Debug("storefile addr from filling method", zap.String("path", wS.Cf.FileStorePath))
	file, err := os.OpenFile(wS.Cf.FileStorePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		wS.logger.Error("open db file error")
		return fmt.Errorf("open db file error: %w", err)
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	var raw []byte
	urlDataStr := new(URLDataStruct)
	urlDataStr.UUID = strconv.Itoa(int(wS.uuid))
	urlDataStr.ShrtURL = shrtURL
	urlDataStr.LngURL = lngURL
	raw, err = json.Marshal(urlDataStr)
	if err != nil {
		return fmt.Errorf("marshalling data to db file error: %w", err)
	}
	if _, err := writer.Write(raw); err != nil {
		return fmt.Errorf("writing data to db file error: %w", err)
	}
	if err := writer.WriteByte('\n'); err != nil {
		return fmt.Errorf("making indent in db file error: %w", err)
	}
	return writer.Flush()
}
