package handler

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/cmd/config"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/cmd/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestURLShortenerRequest(t *testing.T) {
	uS := store.NewURLStorage()
	uS.URLStr["abcdefj"] = "http://yandex.practicum.com"
	conf := config.Config{
		ServAddr: ":8080",
		BaseAddr: "http://localhost:8080/",
	}
	wS := NewWorkStruct(uS, &conf)
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

	for i, tc := range testCases {
		_, tc := i, tc
		for k, v := range uS.URLStr {
			tc.shortURL = k
			tc.testURL = v
		}
		t.Run(tc.method, func(t *testing.T) {
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
			err = result.Body.Close()
			require.NoError(t, err)
		})
	}
}

func TestGetlongURLRequest(t *testing.T) {
	uS := store.NewURLStorage()
	uS.URLStr["abcdefj"] = "http://yandex.practicum.com"
	conf := config.Config{
		ServAddr: ":8080",
		BaseAddr: "http://localhost:8080/",
	}
	wS := NewWorkStruct(uS, &conf)
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

	for i, tc := range testCase {
		_, tc := i, tc
		for k, v := range uS.URLStr {
			tc.shortURL = k
			tc.testURL = v
		}
		t.Run(tc.method, func(t *testing.T) {
			tc.target += tc.target + tc.shortURL
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
		})
	}

}

func TestPostlongURLRequestApi(t *testing.T) {
	uS := store.NewURLStorage()
	uS.URLStr["abcdefj"] = "http://yandex.practicum.com"
	conf := config.Config{
		ServAddr: ":8080",
		BaseAddr: "http://localhost:8080/",
	}
	wS := NewWorkStruct(uS, &conf)
	testCase := []struct {
		name                string
		method              string
		fun                 http.HandlerFunc
		expectedCode        int
		target              string
		expectedBody        string
		expectedContentType string
		testURL             string
		shortURL            string
	}{
		{method: http.MethodPost, fun: PostURLApi(uS, conf), expectedCode: http.StatusCreated, target: "/api/shorten", expectedBody: "", testURL: "{\"url\":\"https://practicum.yandex.ru\"}"},
	}

	for i, tc := range testCase {
		_, tc := i, tc
		for k := range uS.URLStr {
			tc.shortURL = k
		}
		t.Run(tc.method, func(t *testing.T) {
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
		})
	}
}
