package store

import (
	"errors"
	"sync"
)

/*
type URLStore interface {
	CreateShortURL(string) (string, error)
	GetLongURL(string) (string, error)
}*/

type URLStorage struct {
	urlStr map[string]string
	mux    *sync.Mutex
}

func NewURLStorage() *URLStorage {
	return &URLStorage{
		urlStr: make(map[string]string),
		mux:    new(sync.Mutex),
	}
}

/*
const letterBytes = "abcdifghijklmnopqrstuvwxyzABCDIFGHIJKLMNOPQRSTUVWXYZ"

func shorting() string {
	b := make([]byte, 8)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}*/

func (uS *URLStorage) AddShortURL(shrtURL, longURL string) bool {
	uS.mux.Lock()
	defer uS.mux.Unlock()
	/*
		if url == "" {
			return "", errors.New("empty url")
		}*/
	_, ok := uS.urlStr[shrtURL]
	if !ok {
		uS.urlStr[shrtURL] = longURL
		return true
	} else {
		return false
	}
	/*
		for key, value := range uS.urlStr {
			if url == value {
				return key, nil
			}
		}

		shrtURL := shorting()
		uS.urlStr[shrtURL] = url
		return shrtURL, nil*/
}

func (uS *URLStorage) GetLongURL(shrtURL string) (string, error) {
	uS.mux.Lock()
	defer uS.mux.Unlock()

	lngURL, ok := uS.urlStr[shrtURL]
	if !ok {
		return "", errors.New("wrong short url")
	}
	return lngURL, nil
}
