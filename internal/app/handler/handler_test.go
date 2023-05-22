package handler

import (
	"io"
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
	testCases := []struct {
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
		{method: http.MethodPost, fun: PostLongURL(uS, conf), expectedCode: http.StatusCreated, target: "/", expectedBody: "", testURL: "http://yandex.practicum.com"},
	}

	for i, tc := range testCases {
		_, tc := i, tc
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, tc.target, strings.NewReader(tc.testURL))
			w := httptest.NewRecorder()
			h := http.HandlerFunc(tc.fun)
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
		{method: http.MethodGet, fun: GetLongURL(uS, "abcdefj"), expectedCode: http.StatusTemporaryRedirect, target: "/", expectedBody: "", testURL: "http://yandex.practicum.com", shortURL: "abcdefj"},
	}

	for i, tc := range testCase {
		_, tc := i, tc
		t.Run(tc.method, func(t *testing.T) {
			tc.target += tc.target + tc.shortURL
			r := httptest.NewRequest(tc.method, tc.target, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(tc.fun)
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
