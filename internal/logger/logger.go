package logger

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/cmd/config"
	"go.uber.org/zap"
)

var Log *zap.Logger = zap.NewNop()

func InitLog(conf *config.Config) error {
	lvl, err := zap.ParseAtomicLevel(conf.LogLevel)
	if err != nil {
		return errors.New("parse logger level err")
	}

	logCfg := zap.NewProductionConfig()
	logCfg.Level = lvl

	zapLogger, err := logCfg.Build()
	if err != nil {
		return errors.New("builging new zap logger error")
	}

	Log = zapLogger
	return nil
}

type (
	RequestData struct {
		URI    string
		method string
	}
	ResponseData struct {
		status int
		size   int
	}
	LoggingRespWrt struct {
		w http.ResponseWriter
		//RequestData  *RequestData
		ResponseData *ResponseData
	}
)

func (lRW *LoggingRespWrt) Write(b []byte) (int, error) {
	size, err := lRW.w.Write(b)
	lRW.ResponseData.size += size
	return size, err
}

func (lRW *LoggingRespWrt) WriteHeader(stCode int) {
	lRW.w.WriteHeader(stCode)
	lRW.ResponseData.status = stCode
}

func LoggingHandler(h http.HandlerFunc) http.HandlerFunc {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &ResponseData{
			status: 0,
			size:   0,
		}

		LgRspWrt := LoggingRespWrt{
			w:            w,
			ResponseData: responseData,
		}

		h.ServeHTTP(LgRspWrt.w, r)
		duration := time.Since(start)

		Log.Debug("incoming request data",
			zap.String("URl", r.RequestURI),
			zap.String("method", r.Method),
			zap.String("status", strconv.Itoa(responseData.status)),
			zap.String("size", strconv.Itoa(responseData.size)),
			zap.String("duration", strconv.Itoa(int(duration.Milliseconds()))),
		)

	}
	return http.HandlerFunc(logFn)
}
