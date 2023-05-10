package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/cmd/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestURLShortenerRequest(t *testing.T) {
	uS := store.NewURLStorage()
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
		{method: http.MethodPost, expectedCode: http.StatusCreated, target: "/", expectedBody: "", testURL: "http://yandex.practicum.com"},
		{method: http.MethodGet, expectedCode: http.StatusTemporaryRedirect, expectedBody: "", testURL: "http://yandex.practicum.com"},
	}

	for i, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, tc.target, strings.NewReader(tc.testURL))
			w := httptest.NewRecorder()
			rt := NewRouter(uS)
			h := http.HandlerFunc(URLShortenerRequest(rt))
			h(w, r)

			result := w.Result()

			assert.Equal(t, tc.expectedCode, result.StatusCode)
			if i == 0 {
				getBody, err := io.ReadAll(result.Body)
				require.NoError(t, err)
				require.NotEqual(t, "", string(getBody), "empty short url")
				words := strings.Split(string(getBody), "/")
				testCases[0].shortURL = words[len(words)-1]
				err = result.Body.Close()
				require.NoError(t, err)
				testCases[1].shortURL = testCases[0].shortURL
				testCases[1].target = "/" + testCases[1].shortURL
			} else {
				t.Logf("short url for get request test is: " + testCases[0].shortURL)
				require.Equal(t, tc.testURL, result.Header.Get("Location"))
			}
		})
	}
}
