package logg

import (
	"fmt"
	"net/http"

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
