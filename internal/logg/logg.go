package logg

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/config"
	"go.uber.org/zap"
)

func InitLog(conf *config.Config) (*zap.Logger, error) {
	var logger *zap.Logger = zap.NewNop()
	lvl, err := zap.ParseAtomicLevel(conf.LogLevel)
	if err != nil {
		return nil, errors.New("parse logger level err")
	}

	logCfg := zap.NewProductionConfig()
	logCfg.Level = lvl

	zapLogger, err := logCfg.Build()
	if err != nil {
		return nil, errors.New("builging new zap logger error")
	}

	logger = zapLogger
	defer logger.Sync()
	return logger, nil
}

type (
	ResponseData struct {
		status int
		size   int
	}
	LoggingRespWrt struct {
		http.ResponseWriter
		ResponseData *ResponseData
	}
)

func (lRW *LoggingRespWrt) Write(b []byte) (int, error) {
	size, err := lRW.ResponseWriter.Write(b)
	lRW.ResponseData.size += size
	return size, err
}

func (lRW *LoggingRespWrt) WriteHeader(stCode int) {
	lRW.ResponseWriter.WriteHeader(stCode)
	lRW.ResponseData.status = stCode
}

func LoggingHandler(h http.HandlerFunc, logger *zap.Logger) http.HandlerFunc {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &ResponseData{
			status: 0,
			size:   0,
		}

		LgRspWrt := LoggingRespWrt{
			ResponseWriter: w,
			ResponseData:   responseData,
		}

		h.ServeHTTP(&LgRspWrt, r)
		duration := time.Since(start)

		logger.Info("incoming request data",
			zap.String("URl", r.RequestURI),
			zap.String("method", r.Method),
			zap.String("status", strconv.Itoa(responseData.status)),
			zap.String("size", strconv.Itoa(responseData.size)),
			zap.String("duration", strconv.Itoa(int(duration.Milliseconds()))),
		)

	}
	return http.HandlerFunc(logFn)
}
