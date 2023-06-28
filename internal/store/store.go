package store

import (
	"context"
	"errors"
	"sync"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/postgresql"
	"go.uber.org/zap"
)

type URLStorage struct {
	urlStr     map[string]string
	mutex      *sync.Mutex
	uuid       int
	dbFileName string
	logger     *zap.Logger
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

func (uS *URLStorage) PostShortURL(shortURL, longURL, userID string, uuid int32) (string, error) {
	uS.mutex.Lock()
	defer uS.mutex.Unlock()
	_, ok := uS.urlStr[shortURL]
	if ok {
		uS.logger.Info("shortURL: " + shortURL + " and longURL: " + uS.urlStr[shortURL])
		return "", errors.New("shortURL is already exist")
	}
	uS.urlStr[shortURL] = longURL
	uS.logger.Debug("storefile addr from post req", zap.String("path", uS.dbFileName))
	return shortURL, nil
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

func (uS *URLStorage) Ping(ctx context.Context) error {
	return nil
}

func (uS *URLStorage) PostURLBatch(btch []postgresql.DBRowStrct, userID string) ([]string, error) {
	return nil, nil
}

type usersURLs struct {
	shortURL string
	longURL  string
}

func (uS *URLStorage) ReturnAllUserReq(ctx context.Context, userID string) (map[string]string, error) {
	return nil, nil
}

func (uS *URLStorage) DeleteUserURLS(ctxPar context.Context, str []postgresql.URLsForDel) error {
	return nil
}
