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
	mutex  sync.Mutex
}

func NewURLStorage() *URLStorage {
	return &URLStorage{
		urlStr: make(map[string]string),
		mutex:  *new(sync.Mutex),
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
	}
*/
func (uS *URLStorage) PostShortURL(shortURL, longURL string) bool {
	uS.mutex.Lock()
	defer uS.mutex.Unlock()
	/*
		if url == "" {
			return "", errors.New("empty url")
		}*/
	_, ok := uS.urlStr[shortURL]
	if !ok && shortURL != "" {
		uS.urlStr[shortURL] = longURL
		return true
	}
	return false
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
	uS.mutex.Lock()
	defer uS.mutex.Unlock()
	lngURL, ok := uS.urlStr[shrtURL]
	if !ok {
		return "", errors.New("wrong short url")
	}
	return lngURL, nil
}
