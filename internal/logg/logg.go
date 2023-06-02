package logg

import (
	"fmt"
	"net/http"
	"time"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/config"
	"go.uber.org/zap"
)

func InitLog(conf *config.Config) (*zap.Logger, error) {
	var logger = zap.NewNop()
	lvl, err := zap.ParseAtomicLevel(conf.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("parse logger level err: %w", err)
	}

	logCfg := zap.NewProductionConfig()
	logCfg.Level = lvl

	zapLogger, err := logCfg.Build()
	if err != nil {
		return nil, fmt.Errorf("builging new zap logger err: %w", err)
	}

	logger = zapLogger
	defer logger.Sync()
	return logger, nil
}

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

func LoggingHandler(h http.Handler, logger *zap.Logger) http.Handler {
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

		logger.Info("incoming request data",
			zap.String("URl", r.RequestURI),
			zap.String("method", r.Method),
			zap.Int("status", responseData.status),
			zap.Int("size", responseData.size),
			zap.Int("duration", int(duration.Milliseconds())),
		)
	})
}
