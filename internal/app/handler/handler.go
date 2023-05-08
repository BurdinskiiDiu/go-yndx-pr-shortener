package handler

import (
	"io"
	"net/http"
	"strconv"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/cmd/store"
)

type Router struct {
	*http.ServeMux
	uS *store.URLStorage
}

func NewRouter(uS *store.URLStorage) *Router {
	rt := &Router{
		ServeMux: http.NewServeMux(),
		uS:       uS,
	}

	rt.HandleFunc("/", http.HandlerFunc(rt.ComRequest).ServeHTTP)
	return rt
}

func (rt *Router) ComRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		defer r.Body.Close()

		content, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		shrtURL, err := rt.uS.CreateShortURL(string(content))
		if err != nil {

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Length", strconv.Itoa(len(shrtURL)))
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shrtURL))
	} else if r.Method == http.MethodGet {
		srtURL := r.URL.Path
		srtURL = srtURL[1:]
		lngURL, err := rt.uS.GetLongURL(srtURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusTemporaryRedirect)
		w.Write([]byte("Location: " + lngURL))
	} else {
		http.Error(w, "bad method", http.StatusBadRequest)
		return
	}
}
