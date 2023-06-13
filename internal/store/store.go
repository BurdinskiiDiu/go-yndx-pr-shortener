package store

import (
	"errors"
	"sync"

	"go.uber.org/zap"
)

type URLStorage struct {
	urlStr     map[string]string
	mutex      *sync.Mutex
	uuid       int
	dbFileName string
	logger     *zap.Logger
}

func NewURLStorageTest(us *map[string]string, logger *zap.Logger) *URLStorage {
	return &URLStorage{
		urlStr:     *us,
		mutex:      new(sync.Mutex),
		uuid:       0,
		dbFileName: "",
		logger:     logger,
	}
}

func NewURLStorage(logger *zap.Logger) *URLStorage {
	return &URLStorage{
		urlStr:     make(map[string]string),
		mutex:      new(sync.Mutex),
		uuid:       0,
		dbFileName: "",
		logger:     logger,
	}
}

func (uS *URLStorage) PostShortURL(shortURL, longURL string, uuid int32) error {
	uS.mutex.Lock()
	defer uS.mutex.Unlock()
	_, ok := uS.urlStr[shortURL]
	if ok {
		uS.logger.Info("shortURL: " + shortURL + " and longURL: " + uS.urlStr[shortURL])
		return errors.New("this short url is already involved")
	}
	uS.urlStr[shortURL] = longURL
	uS.logger.Debug("storefile addr from post req", zap.String("path", uS.dbFileName))
	return nil
}

func (uS *URLStorage) GetLongURL(shrtURL string) (string, error) {
	uS.mutex.Lock()
	defer uS.mutex.Unlock()
	lngURL, ok := uS.urlStr[shrtURL]
	if !ok {
		return "", errors.New("wrong short url")
	}
	return lngURL, nil
}

func (uS *URLStorage) Ping() error {
	return nil
}
