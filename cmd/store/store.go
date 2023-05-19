package store

import (
	"errors"
	"sync"
)

type URLStorage struct {
	UrlStr map[string]string
	mutex  *sync.Mutex
}

func NewURLStorage() *URLStorage {
	return &URLStorage{
		UrlStr: make(map[string]string),
		mutex:  new(sync.Mutex),
	}
}

func (uS *URLStorage) PostShortURL(shortURL, longURL string) error {
	uS.mutex.Lock()
	defer uS.mutex.Unlock()
	_, ok := uS.UrlStr[shortURL]
	if ok {
		return errors.New("this short url is already involved")
	}
	uS.UrlStr[shortURL] = longURL
	return nil

}

func (uS *URLStorage) GetLongURL(shrtURL string) (string, error) {
	uS.mutex.Lock()
	defer uS.mutex.Unlock()
	lngURL, ok := uS.UrlStr[shrtURL]
	if !ok {
		return "", errors.New("wrong short url")
	}
	return lngURL, nil
}
