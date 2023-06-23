package handler

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/config"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/gzp"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/postgresql"
	"go.uber.org/zap"
)

type URLStore interface {
	PostShortURL(shURL string, lnURL string, uuid int32) (string, error)
	GetLongURL(shURL string) (string, error)
	PostURLBatch([]postgresql.DBRowStrct) ([]string, error)
	PrintlAllDB()
	Ping() error
}

const letterBytes = "abcdifghijklmnopqrstuvwxyzABCDIFGHIJKLMNOPQRSTUVWXYZ"

func shorting() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	fmt.Println(uuid)
	return uuid
}

type URLReq struct {
	URL string `json:"url"`
}

type URLResp struct {
	Result string `json:"result"`
}

type Handlers struct {
	US        URLStore
	Cf        *config.Config
	logger    *zap.Logger
	uuid      int32
	shortings map[string]int
}

func NewHandlers(uS URLStore, cf *config.Config, logger *zap.Logger) *Handlers {
	return &Handlers{
		US:        uS,
		Cf:        cf,
		logger:    logger,
		uuid:      0,
		shortings: make(map[string]int),
	}
}

func (hn *Handlers) CreateShortURL(longURL string) (shrtURL string, err error) {

	cntr := 0

	hn.uuid++
	for cntr < 100 {
		shrtURL = shorting()
		if shrtURL, err = hn.US.PostShortURL(shrtURL, longURL, hn.uuid); err != nil {
			hn.logger.Info(err.Error())
			if strings.Contains(err.Error(), "shortURL is already exist") {
				cntr++
				continue
			} else if strings.Contains(err.Error(), "longURL is already exist") {
				break
			} else {
				break
			}
		}
		if errFF := hn.FileFilling(shrtURL, longURL); errFF != nil {
			hn.logger.Error("createShortURL method, err while filling file", zap.Error(errFF))
		}
		break
	}
	return
}

func (hn *Handlers) PostLongURL() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hn.logger.Info("start post request")
		content, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		longURL := string(content)
		hn.logger.Info("got post message" + longURL)
		var shrtURL string
		var chndStatus bool
		shrtURL, err = hn.CreateShortURL(longURL)
		if err != nil {
			if strings.Contains(err.Error(), "longURL is already exist") {
				chndStatus = true
			} else {
				hn.logger.Error("postLongURL handler error", zap.Error(err))
				return
			}
		}

		bodyResp := hn.Cf.BaseAddr + "/" + shrtURL
		hn.logger.Info("response body message", zap.String("body", bodyResp))
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		if chndStatus {
			hn.logger.Info("chndStatus is true ")
			w.WriteHeader(http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusCreated)
		}
		w.Write([]byte(bodyResp))
	})
}

func (hn *Handlers) GetLongURL(srtURL string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hn.logger.Info("start get request")
		hn.US.PrintlAllDB()
		hn.logger.Info("shortURL is:", zap.String("shortURL", srtURL))
		lngURL, err := hn.US.GetLongURL(srtURL)
		hn.logger. /*Debug*/ Info("longURL is:", zap.String("longURL", lngURL))
		if err != nil {
			hn.logger.Error("getLongURL handler, error while getting long url from store", zap.Error(err))
			return
		}
		hn.logger.Info("response:", zap.String("lngURL", lngURL))
		w.Header().Set("Location", lngURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
		hn.logger.Debug("longURL is:", zap.String("longURL", lngURL))
		w.Write([]byte(lngURL))
		hn.US.PrintlAllDB()
	})
}

func (hn *Handlers) PostURLApi() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hn.logger.Info("start post url api request")
		hn.US.PrintlAllDB()
		var buf bytes.Buffer
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			hn.logger.Error("posrURLApi handler, read from request body err", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		hn.logger.Debug("got postApi message", zap.String("body", buf.String()))

		var urlReq URLReq
		if err := json.Unmarshal(buf.Bytes(), &urlReq); err != nil {
			hn.logger.Error("postURLApi handler, unmarshal func err", zap.Error(err))
			return
		}
		hn.logger.Debug("unmarshaled url from postApi message", zap.String("longURL", urlReq.URL))
		var chndStatus bool
		shrtURL, err := hn.CreateShortURL(urlReq.URL)
		if err != nil {
			if strings.Contains(err.Error(), "longURL is already exist") {
				chndStatus = true
			} else {
				hn.logger.Error("postLongURL handler error", zap.Error(err))
				return
			}
		}

		hn.logger.Debug("short url", zap.String("shortURL", shrtURL))
		var urlResp URLResp
		urlResp.Result = hn.Cf.BaseAddr + "/" + shrtURL
		resp, err := json.Marshal(urlResp)
		if err != nil {
			hn.logger.Error("postURLApi handler, marshal func error", zap.Error(err))
			return
		}

		hn.logger.Debug("response for postApi request", zap.String("response", string(resp)))
		w.Header().Set("Content-Type", "application/json")
		if chndStatus {
			w.WriteHeader(http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusCreated)
		}
		w.Write(resp)
		hn.US.PrintlAllDB()
	})
}

// log middleware
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

func (hn *Handlers) LoggingHandler(h http.Handler) http.Handler {
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

		hn.logger.Info("incoming request data",
			zap.String("URl", r.RequestURI),
			zap.String("method", r.Method),
			zap.Int("status", responseData.status),
			zap.Int("size", responseData.size),
			zap.Int("duration", int(duration.Milliseconds())),
		)
	})
}

// gzip middleware
func (hn *Handlers) GZipMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		accptEnc := r.Header.Get("Accept-Encoding")
		hn.logger.Debug("acceptEnc", zap.String("accptEnc", accptEnc))
		contType := r.Header.Get("Content-Type")
		hn.logger.Debug("contType", zap.String("contType", contType))
		suppGZip := strings.Contains(accptEnc, "gzip")
		if suppGZip {
			cw := gzp.NewCompressWriter(w)
			ow = cw
		}
		hn.logger.Debug("acceptEnc", zap.String("accptEnc", accptEnc))
		cntntEnc := r.Header.Get("Content-Encoding")
		hn.logger.Debug("cntntEnc", zap.String("cntntEnc", cntntEnc))
		sendGZip := strings.Contains(cntntEnc, "gzip")
		if sendGZip {
			cr, err := gzp.NewCompressReader(r.Body)
			if err != nil {
				hn.logger.Error("compersReader creation err", zap.Error(err))
				return
			}
			r.Body = cr
			defer cr.Close()
		}
		hn.logger.Debug("response", zap.String("response", r.RequestURI))

		h.ServeHTTP(ow, r)
	})
}

// db handler
func (hn *Handlers) GetDBPing() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := hn.US.Ping(); err != nil {
			hn.logger.Error("getDBping handler error", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}

// workStruct init
type URLDataStruct struct {
	UUID    string `json:"uuid"`
	ShrtURL string `json:"short_url"`
	LngURL  string `json:"original_url"`
}

func (hn *Handlers) GetStoreBackup() error {
	hn.logger.Debug("storefile addr from createfile", zap.String("path", hn.Cf.FileStorePath))
	file, err := os.OpenFile(hn.Cf.FileStorePath, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		hn.logger.Error("open storeFile error")
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
			hn.logger.Error("unmarhalling store_file error")
			return err
		}
		uuid, err := strconv.Atoi(urlDataStr.UUID)
		if err != nil {
			return errors.New("error while filling db from backup file, uuid conv to int err:" + err.Error())
		}
		_, err = hn.US.PostShortURL(urlDataStr.ShrtURL, urlDataStr.LngURL, int32(uuid))
		if err != nil {
			hn.logger.Error("getStoreBackup error, try to write itno db", zap.Error(err))
		}
	}

	if urlDataStr.UUID != "" {
		uuid, err := strconv.Atoi(urlDataStr.UUID)
		if err != nil {
			hn.logger.Error("gettitng last uuid error, file is damaged")
			return err
		}
		hn.uuid = int32(uuid)
	}
	return nil
}

func (hn *Handlers) FileFilling(shrtURL, lngURL string) error {
	hn.logger.Debug("storefile addr from filling method", zap.String("path", hn.Cf.FileStorePath))
	file, err := os.OpenFile(hn.Cf.FileStorePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		hn.logger.Error("open db file error")
		return fmt.Errorf("open db file error: %w", err)
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	var raw []byte
	urlDataStr := new(URLDataStruct)
	urlDataStr.UUID = strconv.Itoa(int(hn.uuid))
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

// PostBatch handler
type batchReqStruct struct {
	CorrID  string `json:"correlation_id"`
	OrigURL string `json:"original_url"`
}

type batchRespStruct struct {
	CorrID   string `json:"correlation_id"`
	ShortURL string `json:"short_url"`
}

func (hn *Handlers) PostBatch2() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var buf bytes.Buffer
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			hn.logger.Error("PostBatch handler, read from request body err", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		str := buf.String()
		cnt := strings.Count(str, "correlation_id")
		hn.logger.Info("cnt of json rows", zap.Int("cnt", cnt))

		urlReq := make([]batchReqStruct, cnt)
		if err := json.Unmarshal(buf.Bytes(), &urlReq); err != nil {
			hn.logger.Error("PostBatch handler, unmarshal func err", zap.Error(err))
			return
		}

		fmt.Println(urlReq)
		urlResp := make([]batchRespStruct, cnt)
		for i, v := range urlReq {
			urlResp[i].CorrID = v.CorrID
			shortURL, err := hn.CreateShortURL(v.OrigURL)
			hn.logger.Info("shortURL is " + shortURL)
			if err != nil {
				hn.logger.Error("PostBatch handler, creating short url err", zap.Error(err))
				return
			}
			urlResp[i].ShortURL = hn.Cf.BaseAddr + "/" + shortURL
			hn.logger.Info("result short URL " + urlResp[i].ShortURL)
			hn.logger.Info("added is successful, add № is " + strconv.Itoa(i))
		}

		resp, err := json.Marshal(urlResp)
		if err != nil {
			hn.logger.Error("PostBatch handler, marshal func error", zap.Error(err))
			return
		}

		hn.logger.Info("response for postApi request", zap.String("response", string(resp)))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(resp)
	})
}

func (hn *Handlers) PostBatch() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hn.US.PrintlAllDB()
		var buf bytes.Buffer
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			hn.logger.Error("PostBatch handler, read from request body err", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		str := buf.String()
		cnt := strings.Count(str, "correlation_id")
		hn.logger.Info("cnt of json rows", zap.Int("cnt", cnt))

		urlReq := make([]batchReqStruct, cnt)
		if err := json.Unmarshal(buf.Bytes(), &urlReq); err != nil {
			hn.logger.Error("PostBatch handler, unmarshal func err", zap.Error(err))
			return
		}
		fmt.Println(urlReq)
		btchStr := make([]postgresql.DBRowStrct, 0)
		var btchRow postgresql.DBRowStrct

		urlResp := make([]batchRespStruct, cnt)
		for i, v := range urlReq {
			urlResp[i].CorrID = v.CorrID
			hn.uuid++
			btchRow.ID = int(hn.uuid)
			btchRow.LongURL = v.OrigURL
			shortURL := shorting()
			btchRow.ShortURL = shortURL
			//urlResp[i].ShortURL = shortURL
			btchStr = append(btchStr, btchRow)
			//shortURL, err := hn.CreateShortURL(v.OrigURL)
			//hn.logger.Info("shortURL is " + shortURL)
			//if err != nil {
			//hn.logger.Error("PostBatch handler, creating short url err", zap.Error(err))
			//return
			//}

			/*if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates") {
				hn.logger.Info(" created shrtURL", zap.String("shrtURL", shortURL))
				shortURL, err = hn.US.GetShortURL(v.OrigURL)
				if err != nil {
					hn.logger.Error("getting already existed short url error", zap.Error(err))
					return
				}
				hn.logger.Info("existed shrtURL", zap.String("shrtURL", shortURL))
			} else {
				hn.logger.Error("PostBatch handler, creating short url err", zap.Error(err))
				return
			}
			/*hn.logger.Error("PostBatch handler, creating short url err", zap.Error(err))
			return*/
			/*}*/
			//urlResp[i].ShortURL = hn.Cf.BaseAddr + "/" + shortURL
			//hn.logger.Info("result short URL " + urlResp[i].ShortURL)
			//hn.logger.Info("added is successful, add № is " + strconv.Itoa(i))
		}
		retShrtURL, err := hn.US.PostURLBatch(btchStr)
		if err != nil {
			hn.logger.Error("post batch error", zap.Error(err))
			return
		}

		for i, v := range retShrtURL {
			urlResp[i].ShortURL = hn.Cf.BaseAddr + "/" + v
		}

		resp, err := json.Marshal(urlResp)
		if err != nil {
			hn.logger.Error("PostBatch handler, marshal func error", zap.Error(err))
			return
		}
		hn.logger.Info("response for postApi request", zap.String("response", string(resp)))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(resp)
		hn.US.PrintlAllDB()
	})
}
