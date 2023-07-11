package store

import (
	"sync"

	"go.uber.org/zap"
)

func NewURLStorageTest(us *map[string]UlStr, logger *zap.Logger) *URLStorage {
	return &URLStorage{
		urlStr:     *us,
		mutex:      new(sync.Mutex),
		uuid:       0,
		dbFileName: "",
		logger:     logger,
	}
}

func NewUlStr() *UlStr {
	return &UlStr{
		ID:      0,
		UserID:  "",
		LongURL: "http://yandex.practicum.com",
	}
}
