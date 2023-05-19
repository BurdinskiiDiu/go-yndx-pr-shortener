package store

import (
	"errors"
	"sync"
)

type URLStorage struct {
	urlStr map[string]string
	mutex  *sync.Mutex
}

func NewURLStorage() *URLStorage {
	return &URLStorage{
		urlStr: make(map[string]string),
		mutex:  new(sync.Mutex),
	}
}

func (uS *URLStorage) PostShortURL(shortURL, longURL string) error {
	uS.mutex.Lock()
	defer uS.mutex.Unlock()
	_, ok := uS.urlStr[shortURL]
	if ok {
		return errors.New("this short url is already involved")
	}
	uS.urlStr[shortURL] = longURL
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
