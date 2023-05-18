package store

import (
	"errors"
	"sync"
)

type URLStorage struct {
	urlStr map[string]string
	mutex  sync.Mutex
}

func NewURLStorage() *URLStorage {
	return &URLStorage{
		urlStr: make(map[string]string),
		mutex:  *new(sync.Mutex),
	}
}

func (uS *URLStorage) PostShortURL(shortURL, longURL string) bool {
	uS.mutex.Lock()
	defer uS.mutex.Unlock()
	_, ok := uS.urlStr[shortURL]
	if !ok && shortURL != "" {
		uS.urlStr[shortURL] = longURL
		return true
	}
	return false
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
