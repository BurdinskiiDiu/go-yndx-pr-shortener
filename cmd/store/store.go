package store

import (
	"errors"
	"math/rand"
	"sync"
)

type UrlStorage struct {
	urlStr map[string]string
	sync.Mutex
}

func NewUrlStorage() *UrlStorage {
	return &UrlStorage{
		urlStr: make(map[string]string),
	}
}

const letterBytes = "abcdifghijklmnopqrstuvwxyzABCDIFGHIJKLMNOPQRSTUVWXYZ"

func shorting() string {
	b := make([]byte, 8)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func (uS *UrlStorage) CreateShortUrl(url string) (string, error) {
	uS.Lock()
	defer uS.Unlock()

	if url == "" {
		return "", errors.New("empty url")
	}

	for key, value := range uS.urlStr {
		if url == value {
			return key, nil
		}
	}

	shrtUrl := shorting()
	uS.urlStr[shrtUrl] = url
	return shrtUrl, nil
}

func (us *UrlStorage) GetLongUrl(shrtUrl string) (string, error) {
	lngUrl, ok := us.urlStr[shrtUrl]
	if !ok {
		return "", errors.New("wrong short url")
	} else {
		return lngUrl, nil
	}
}
