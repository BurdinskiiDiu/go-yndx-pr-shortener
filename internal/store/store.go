package store

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/postgresql"
	"go.uber.org/zap"
)

type UlStr struct {
	Id      int32
	User_id string
	LongURL string
}

type URLStorage struct {
	urlStr     map[string]UlStr
	mutex      *sync.Mutex
	uuid       int
	dbFileName string
	logger     *zap.Logger
}

func NewURLStorage(logger *zap.Logger) *URLStorage {
	return &URLStorage{
		urlStr:     make(map[string]UlStr),
		mutex:      new(sync.Mutex),
		uuid:       0,
		dbFileName: "",
		logger:     logger,
	}
}

func (uS *URLStorage) PostShortURL(shortURL, longURL, userID string, uuid int32) (string, error) {
	uS.mutex.Lock()
	defer uS.mutex.Unlock()
	var ulStr UlStr
	ulStr.User_id = userID
	ulStr.LongURL = longURL
	ulStr.Id = uuid
	_, ok := uS.urlStr[shortURL]
	if ok {
		uS.logger.Info("shortURL: " + shortURL + " and longURL: " + uS.urlStr[shortURL].LongURL)
		return "", errors.New("shortURL is already exist")
	}
	uS.urlStr[shortURL] = ulStr
	uS.logger.Debug("storefile addr from post req", zap.String("path", uS.dbFileName))
	return shortURL, nil
}

func (uS *URLStorage) GetLongURL(shrtURL string) (string, error) {
	uS.mutex.Lock()
	defer uS.mutex.Unlock()
	ulStr, ok := uS.urlStr[shrtURL]
	if !ok {
		return "", errors.New("wrong short url")
	}
	return ulStr.LongURL, nil
}

func (uS *URLStorage) Ping(ctx context.Context) error {
	return nil
}

func (uS *URLStorage) PostURLBatch(btch []postgresql.DBRowStrct, userID string) ([]string, error) {
	shrtURLSlc := make([]string, 0)
	for _, v := range btch {
		shrtURL, err := uS.PostShortURL(v.ShortURL, v.LongURL, userID, int32(v.ID))
		if err != nil {
			return nil, fmt.Errorf("error while postBatching to map: %w", err)
		}
		shrtURLSlc = append(shrtURLSlc, shrtURL)
	}
	return shrtURLSlc, nil
}

type usersURLs struct {
	shortURL string
	longURL  string
}

func (uS *URLStorage) ReturnAllUserReq(ctx context.Context, userID string) (map[string]string, error) {
	ans := make(map[string]string, 0)

	for i, v := range uS.urlStr {
		if v.User_id == userID {
			ans[v.LongURL] = i
		}
	}
	return ans, nil
}

func (uS *URLStorage) DeleteUserURLS(ctxPar context.Context, str []postgresql.URLsForDel) error {
	for _, v := range str {
		_, ok := uS.urlStr[v.ShortURL]
		if ok {
			delete(uS.urlStr, v.ShortURL)
		}
	}
	return nil
}
