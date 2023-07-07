package handler

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/authentication"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/config"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/gzp"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/postgresql"

	"go.uber.org/zap"
)

type ctxUserKey string

var currUser ctxUserKey = "userID"

type URLStore interface {
	PostShortURL(shURL, lnURL, userID string, uuid int32) (string, error)
	GetLongURL(shURL string) (string, error)
	PostURLBatch(tchStr []postgresql.DBRowStrct, userID string) ([]string, error)
	ReturnAllUserReq(ctx context.Context, userID string) (map[string]string, error)
	DeleteUserURLS(ctx context.Context, str []postgresql.URLsForDel) error
	Ping(ctx context.Context) error
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
	US         URLStore
	Cf         *config.Config
	logger     *zap.Logger
	uuid       int32
	forDel     [][]postgresql.URLsForDel
	inpURLSChn chan []postgresql.URLsForDel
	delMtx     *sync.Mutex
}

func NewHandlers(uS URLStore, inpURLSChn chan []postgresql.URLsForDel, cf *config.Config, logger *zap.Logger, delMtx *sync.Mutex) *Handlers {
	return &Handlers{
		US:         uS,
		Cf:         cf,
		logger:     logger,
		uuid:       0,
		forDel:     make([][]postgresql.URLsForDel, 0),
		inpURLSChn: inpURLSChn,
		delMtx:     delMtx,
	}
}

func (hn *Handlers) CreateShortURL(longURL, userID string) (shrtURL string, err error) {

	cntr := 0
	hn.uuid++
	for cntr < 100 {
		shrtURL = shorting()
		if shrtURL, err = hn.US.PostShortURL(shrtURL, longURL, userID, hn.uuid); err != nil {
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
		if errFF := hn.FileFilling(shrtURL, longURL, userID); errFF != nil {
			hn.logger.Error("createShortURL method, err while filling file", zap.Error(errFF))
		}
		break
	}
	return
}

func (hn *Handlers) PostLongURL() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hn.logger.Debug("start PostLongURL")
		content, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		longURL := string(content)
		hn.logger.Debug("got post message" + longURL)
		var shrtURL string
		var chndStatus bool

		userID, ok := r.Context().Value(currUser).(string)
		if !ok {
			userID = ""
		}

		shrtURL, err = hn.CreateShortURL(longURL, userID)
		if err != nil {
			if strings.Contains(err.Error(), "longURL is already exist") {
				chndStatus = true
			} else {
				hn.logger.Error("postLongURL handler error", zap.Error(err))
				return
			}
		}

		bodyResp := hn.Cf.BaseAddr + "/" + shrtURL
		hn.logger.Debug("response body message", zap.String("body", bodyResp))
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
		hn.logger.Debug("start GetLongURL")
		hn.logger.Debug("shortURL is:", zap.String("shortURL", srtURL))
		lngURL, err := hn.US.GetLongURL(srtURL)
		hn.logger.Debug("longURL is:", zap.String("longURL", lngURL))
		if err != nil {
			hn.logger.Error("getLongURL handler, error while getting long url from store", zap.Error(err))
			return
		}
		if lngURL == "" {
			w.WriteHeader(http.StatusGone)
		} else {
			w.Header().Set("Location", lngURL)
			w.WriteHeader(http.StatusTemporaryRedirect)
			w.Write([]byte(lngURL))
		}
	})
}

func (hn *Handlers) PostURLApi() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hn.logger.Debug("start PostURLApi")
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

		userID, ok := r.Context().Value(currUser).(string)
		if !ok {
			userID = ""
		}
		shrtURL, err := hn.CreateShortURL(urlReq.URL, userID)
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
		hn.logger.Debug("start LoggingHandler")
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
		ctx := context.TODO()
		if err := hn.US.Ping(ctx); err != nil {
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
	UsrID   string `json:"user_id"`
}

func (hn *Handlers) GetStoreBackup() error {
	hn.logger.Debug("start GetStoreBackup")
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
		_, err = hn.US.PostShortURL(urlDataStr.ShrtURL, urlDataStr.LngURL, urlDataStr.UsrID, int32(uuid))
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

func (hn *Handlers) FileFilling(shrtURL, lngURL, userID string) error {
	hn.logger.Debug("start FileFilling")
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
	urlDataStr.UsrID = userID
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

func (hn *Handlers) PostBatch() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hn.logger.Debug("start PostBatch")
		var buf bytes.Buffer
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			hn.logger.Error("PostBatch handler, read from request body err", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		urlReq := make([]batchReqStruct, 0)
		if err := json.Unmarshal(buf.Bytes(), &urlReq); err != nil {
			hn.logger.Error("PostBatch handler, unmarshal func err", zap.Error(err))
			return
		}
		btchStr := make([]postgresql.DBRowStrct, 0)
		var btchRow postgresql.DBRowStrct
		urlResparr := make([]batchRespStruct, 0)
		var urlResp batchRespStruct
		for _, v := range urlReq {
			urlResp.CorrID = v.CorrID
			hn.uuid++
			btchRow.ID = int(hn.uuid)
			btchRow.LongURL = v.OrigURL
			shortURL := shorting()
			btchRow.ShortURL = shortURL
			btchStr = append(btchStr, btchRow)
			urlResparr = append(urlResparr, urlResp)
		}
		userID, ok := r.Context().Value(currUser).(string)
		if !ok {
			userID = ""
		}
		retShrtURL, err := hn.US.PostURLBatch(btchStr, userID)
		if err != nil {
			hn.logger.Error("post batch error", zap.Error(err))
			return
		}

		for i, v := range retShrtURL {
			urlResparr[i].ShortURL = hn.Cf.BaseAddr + "/" + v
		}

		resp, err := json.Marshal(urlResparr)
		if err != nil {
			hn.logger.Error("PostBatch handler, marshal func error", zap.Error(err))
			return
		}
		hn.logger.Debug("response for postApi request", zap.String("response", string(resp)))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(resp)
	})
}

// authentication middleware
func (hn *Handlers) AuthMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hn.logger.Debug("start AuthMiddleware")
		var noCookie bool
		cookie, err := r.Cookie("authentication")

		if err != nil {
			hn.logger.Info("cookie err, " + err.Error())
			if !errors.Is(err, http.ErrNoCookie) {
				hn.logger.Error("getting request cookie error", zap.Error(err))
				return
			}
			hn.logger.Info("request without necessary cookie")
			noCookie = true
		}
		var createCookie bool
		var userID, cookieStr string
		if !noCookie {
			hn.logger.Debug("gotted cookie string is", zap.String("hexcookie", cookie.Value))
			gCookieStr, err := url.QueryUnescape(cookie.Value)
			if err != nil {
				hn.logger.Error("decoding cookie string error", zap.Error(err))
				return
			}

			userID, err = authentication.CheckCookie(gCookieStr, *hn.Cf)
			if err != nil {
				createCookie = true
			}

		} else {
			createCookie = true
		}

		if createCookie {

			userID, cookieStr, err = authentication.CreateUserID(*hn.Cf)
			if err != nil {
				hn.logger.Error("creating user id error", zap.Error(err))
				return
			}

			respCookieStr := url.QueryEscape(cookieStr)
			respCookie := http.Cookie{
				Name:    "authentication",
				Value:   respCookieStr,
				Expires: time.Now().Add(1 * time.Hour),
			}
			http.SetCookie(w, &respCookie)
		}
		ctx := context.WithValue(r.Context(), currUser, userID)

		if noCookie && r.Method == http.MethodGet && r.URL.Path == "/api/user/urls" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

type UsersURLs struct {
	ShortURL string `json:"short_url"`
	LongURL  string `json:"original_url"`
}

// GetUsersURLs handler
func (hn *Handlers) GetUsersURLs() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hn.logger.Debug("start GetUsersURLs")

		ctx := r.Context()
		userID := ctx.Value(currUser)
		ans, err := hn.US.ReturnAllUserReq(ctx, userID.(string))
		if err != nil {
			hn.logger.Error("getUsersURLs error", zap.Error(err))
			return
		}
		if len(ans) == 0 {
			w.WriteHeader(http.StatusNoContent)
		} else {
			var usersURLs UsersURLs
			usrURLsArr := make([]UsersURLs, 0)
			for i, v := range ans {
				v = hn.Cf.BaseAddr + "/" + v
				usersURLs.LongURL = i
				usersURLs.ShortURL = v
				usrURLsArr = append(usrURLsArr, usersURLs)
			}

			resp, err := json.Marshal(&usrURLsArr)
			if err != nil {
				hn.logger.Error("getUsersURLs, error while marshalling response data", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(resp)
		}
	})
}

func (hn *Handlers) DeleteUsersURLs() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hn.logger.Debug("start DeleteUsersURLs")
		var buf bytes.Buffer
		_, err := buf.ReadFrom(r.Body)
		var delURLstr postgresql.URLsForDel
		if err != nil {
			hn.logger.Error("DeleteUsersURLs handler, read from request body err", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		urlsSlc := make([]string, 0)
		if err := json.Unmarshal(buf.Bytes(), &urlsSlc); err != nil {
			hn.logger.Error("DeleteUsersURLs handler, unmarshal func err", zap.Error(err))
			return
		}

		userID, ok := r.Context().Value(currUser).(string)
		if !ok {
			userID = ""
		}
		delURLsSlc := make([]postgresql.URLsForDel, 0)
		for _, v := range urlsSlc {
			delURLstr.UserID = userID
			delURLstr.ShortURL = v
			delURLsSlc = append(delURLsSlc, delURLstr)
		}
		//hn.inpURLSChn <- delURLsSlc

		hn.forDel = append(hn.forDel, delURLsSlc)
		w.WriteHeader(http.StatusAccepted)
		/*
			ctx := context.TODO()
			go func() {
				hn.US.DeleteUserURLS(ctx, delURLsSlc)
				if err != nil {
					hn.logger.Error("async deleting userURLS err", zap.Error(err))
				}
			}()
		*/
	})
}

func (hn *Handlers) DelURLSBatch() {
	ctx := context.TODO()
	ticker := time.NewTicker(5 * time.Second)
	delURL := make([]postgresql.URLsForDel, 0)

	for {
		select {
		case delURL = <-hn.inpURLSChn:
			err := hn.US.DeleteUserURLS(ctx, delURL)
			if err != nil {
				hn.logger.Debug("error while del urls:" + err.Error())
				continue
			}
		case <-ticker.C:
			hn.delMtx.Lock()
			if len(hn.forDel) == 0 {
				hn.delMtx.Unlock()
				continue
			}
			if len(hn.forDel) <= hn.Cf.DelChnlLen {
				for _, v := range hn.forDel {
					hn.inpURLSChn <- v
				}
				hn.forDel = nil
				hn.delMtx.Unlock()
				continue
			}
			for i := 0; i < hn.Cf.DelChnlLen; i++ {
				hn.inpURLSChn <- hn.forDel[i]
			}
			hn.forDel = hn.forDel[hn.Cf.DelChnlLen-1:]
			hn.delMtx.Unlock()
		}
	}
}
