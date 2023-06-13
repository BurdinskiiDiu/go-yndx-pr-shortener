package handler

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/config"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/logg"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/postgresql"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestURLShortenerRequest(t *testing.T) {
	conf := new(config.Config)
	conf.DBdsn = ""
	conf.FileStorePath = "/tmp/short-url-db.json"
	logger, err := logg.InitLog(conf)
	urlStr := make(map[string]string)
	urlStr["abcdefj"] = "http://yandex.practicum.com"
	uS := store.NewURLStorageTest(&urlStr, logger)

	conf.ServAddr = ":8080"
	conf.BaseAddr = "http://localhost:8080/"

	ctx := context.TODO()
	db := postgresql.NewClientDBStruct(ctx, logger, conf)
	wS := NewWorkStruct(uS, conf, logger, db, ctx)

	if err != nil {
		log.Fatal(err)
	}
	testCases := []struct {
		name                string
		method              string
		expectedCode        int
		target              string
		expectedBody        string
		expectedContentType string
		testURL             string
		shortURL            string
	}{
		{method: http.MethodPost, expectedCode: http.StatusCreated, target: "/", expectedBody: ""},
	}

	for _, tc := range testCases {
		tc := tc
		for k, v := range urlStr {
			tc.shortURL = k
			tc.testURL = v
		}
		log.Println("post req shortURL is: " + tc.shortURL)
		t.Run(tc.method, func(t *testing.T) {
			log.Println("1st test start")
			r := httptest.NewRequest(tc.method, tc.target, strings.NewReader(tc.testURL))
			w := httptest.NewRecorder()
			h := http.HandlerFunc(wS.PostLongURL())
			h(w, r)

			result := w.Result()
			assert.Equal(t, tc.expectedCode, result.StatusCode)
			getBody, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			require.NotEqual(t, "", string(getBody), "empty short url")
			words := strings.Split(string(getBody), "/")
			tc.shortURL = words[len(words)-1]
			log.Println("post req shortURL is: " + tc.shortURL)
			err = result.Body.Close()
			require.NoError(t, err)
			log.Println("1st test finish")
		})
	}
}

func TestGetlongURLRequest(t *testing.T) {
	conf := new(config.Config)
	conf.DBdsn = ""
	conf.FileStorePath = "/tmp/short-url-db.json"
	logger, err := logg.InitLog(conf)
	urlStr := make(map[string]string)
	urlStr["abcdefj"] = "http://yandex.practicum.com"
	uS := store.NewURLStorageTest(&urlStr, logger)

	conf.ServAddr = ":8080"
	conf.BaseAddr = "http://localhost:8080/"

	ctx := context.TODO()
	db := postgresql.NewClientDBStruct(ctx, logger, conf)
	wS := NewWorkStruct(uS, conf, logger, db, ctx)

	if err != nil {
		log.Fatal(err)
	}
	testCase := []struct {
		name                string
		method              string
		expectedCode        int
		target              string
		expectedBody        string
		expectedContentType string
		testURL             string
		shortURL            string
	}{
		{method: http.MethodGet, expectedCode: http.StatusTemporaryRedirect, target: "/", expectedBody: ""},
	}

	for _, tc := range testCase {
		tc := tc
		for k, v := range urlStr {
			tc.shortURL = k
			tc.testURL = v
		}
		log.Println("testURL is: " + tc.testURL)
		t.Run(tc.method, func(t *testing.T) {
			log.Println("2nd test start")
			log.Println("get req shortURL is: " + tc.shortURL)
			tc.target += tc.target + tc.shortURL
			log.Println("post req target is: " + tc.target)
			r := httptest.NewRequest(tc.method, tc.target, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(wS.GetLongURL(tc.shortURL))
			h(w, r)

			result := w.Result()
			assert.Equal(t, tc.expectedCode, result.StatusCode)
			t.Logf("short url for get request test is: " + tc.shortURL)
			require.Equal(t, tc.testURL, result.Header.Get("Location"))
			err := result.Body.Close()
			require.NoError(t, err)
			log.Println("2nd test finish")
		})
	}

}

func TestPostlongURLRequestApi(t *testing.T) {
	conf := new(config.Config)
	conf.DBdsn = ""
	conf.FileStorePath = "/tmp/short-url-db.json"
	logger, err := logg.InitLog(conf)
	urlStr := make(map[string]string)
	urlStr["abcdefj"] = "http://yandex.practicum.com"
	uS := store.NewURLStorageTest(&urlStr, logger)

	conf.ServAddr = ":8080"
	conf.BaseAddr = "http://localhost:8080/"

	ctx := context.TODO()
	db := postgresql.NewClientDBStruct(ctx, logger, conf)
	wS := NewWorkStruct(uS, conf, logger, db, ctx)

	if err != nil {
		log.Fatal(err)
	}
	testCase := []struct {
		name                string
		method              string
		expectedCode        int
		target              string
		expectedBody        string
		expectedContentType string
		testURL             string
		shortURL            string
	}{
		{method: http.MethodPost, expectedCode: http.StatusCreated, target: "/api/shorten", expectedBody: "", testURL: "{\"url\":\"https://practicum.yandex.ru\"}"},
	}

	for _, tc := range testCase {
		tc := tc
		for k := range urlStr {
			tc.shortURL = k
		}
		t.Run(tc.method, func(t *testing.T) {
			log.Println("3d test start")
			r := httptest.NewRequest(tc.method, tc.target, strings.NewReader(tc.testURL))
			w := httptest.NewRecorder()
			h := http.HandlerFunc(wS.PostURLApi())
			h(w, r)

			result := w.Result()
			assert.Equal(t, tc.expectedCode, result.StatusCode)
			log.Println(result.Body)
			getBody, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			require.NotEqual(t, "", string(getBody), "empty short url")
			words := strings.Split(string(getBody), "/")
			tc.shortURL = words[len(words)-1]
			err = result.Body.Close()
			require.NoError(t, err)
			log.Println("3nd test finish")
		})
	}
}
