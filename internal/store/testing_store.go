package store

import (
	"sync"

	"go.uber.org/zap"
)

func NewURLStorageTest(us *map[string]string, logger *zap.Logger) *URLStorage {
	return &URLStorage{
		urlStr:     *us,
		mutex:      new(sync.Mutex),
		uuid:       0,
		dbFileName: "",
		logger:     logger,
	}
}
