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
		{method: http.MethodGet, fun: GetLongURL(uS, "abcdefj"), expectedCode: http.StatusTemporaryRedirect, expectedBody: "", testURL: "http://yandex.practicum.com"},
	}

	for i, tc := range testCases {
		i, tc := i, tc
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, tc.target, strings.NewReader(tc.testURL))
			w := httptest.NewRecorder()
			h := http.HandlerFunc(tc.fun)
			h(w, r)

			result := w.Result()
			assert.Equal(t, tc.expectedCode, result.StatusCode)
			if i != 0 {
				t.Logf("short url for get request test is: " + testCases[0].shortURL)
				require.Equal(t, tc.testURL, result.Header.Get("Location"))
				return
			}
			getBody, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			require.NotEqual(t, "", string(getBody), "empty short url")
			words := strings.Split(string(getBody), "/")
			testCases[0].shortURL = words[len(words)-1]
			err = result.Body.Close()
			require.NoError(t, err)
			testCases[1].shortURL = testCases[0].shortURL
			testCases[1].target = "/" + testCases[1].shortURL
		})
	}
}
