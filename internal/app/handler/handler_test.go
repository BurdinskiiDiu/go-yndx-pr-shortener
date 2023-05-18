package handler

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/cmd/config"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/cmd/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var uS = store.NewURLStorage()
var conf = config.Config{
	ServAddr: ":8080",
	BaseAddr: "http://localhost:8080/",
}
var mux = new(sync.Mutex)

type testCase struct {
	name                string
	method              string
	fun                 http.HandlerFunc
	expectedCode        int
	target              string
	expectedBody        string
	expectedContentType string
	testURL             string
	shortURL            string
}

var testCasePost = testCase{method: http.MethodPost, fun: PostLongURL(uS, conf), expectedCode: http.StatusCreated, target: "/", expectedBody: "", testURL: "http://yandex.practicum.com"}

//var testCaseGet = testCase{method: http.MethodGet, fun: GetLongURL(uS), expectedCode: http.StatusTemporaryRedirect, expectedBody: "", testURL: "http://yandex.practicum.com"}

func TestURLPostRequest(t *testing.T) {
	mux.Lock()
	defer mux.Unlock()
	t.Run(testCasePost.method, func(t *testing.T) {
		r := httptest.NewRequest(testCasePost.method, testCasePost.target, strings.NewReader(testCasePost.testURL))
		w := httptest.NewRecorder()
		//rt := NewRouter(uS)
		h := http.HandlerFunc(testCasePost.fun)
		h(w, r)
		result := w.Result()
		assert.Equal(t, testCasePost.expectedCode, result.StatusCode)
		getBody, err := io.ReadAll(result.Body)
		require.NoError(t, err)
		require.NotEqual(t, "", string(getBody), "empty short url")
		words := strings.Split(string(getBody), "/")
		testCasePost.shortURL = words[len(words)-1]
		err = result.Body.Close()
		require.NoError(t, err)
		//testCaseGet.shortURL = testCasePost.shortURL
		//testCaseGet.target = "/" + testCaseGet.shortURL
		log.Printf(testCasePost.shortURL)
	})
}

func TestURLGetRequest(t *testing.T) {
	mux.Lock()
	defer mux.Unlock()
	t.Run(http.MethodGet, func(t *testing.T) {
		testCasePost.target += testCasePost.shortURL
		t.Logf("target for get request test is: " + testCasePost.target)
		r := httptest.NewRequest(http.MethodGet, testCasePost.target, strings.NewReader(testCasePost.testURL))
		w := httptest.NewRecorder()
		//rt := NewRouter(uS)
		h := http.HandlerFunc(testCasePost.fun)
		h(w, r)

		result := w.Result()
		assert.Equal(t, testCasePost.expectedCode, result.StatusCode)

		t.Logf("short url for get request test is: " + testCasePost.shortURL)
		require.Equal(t, testCasePost.testURL, result.Header.Get("Location"))
	})
}

/*func TestURLShortenerRequest(t *testing.T) {
	uS := store.NewURLStorage()
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
		{method: http.MethodGet, fun: GetLongURL(uS), expectedCode: http.StatusTemporaryRedirect, expectedBody: "", testURL: "http://yandex.practicum.com"},
	}

	for i, tc := range testCases {
		i, tc := i, tc
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, tc.target, strings.NewReader(tc.testURL))
			w := httptest.NewRecorder()
			//rt := NewRouter(uS)
			h := http.HandlerFunc(tc.fun)
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
}*/
