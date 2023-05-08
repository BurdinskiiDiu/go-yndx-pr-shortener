package handler

import (
	"io"
	"net/http"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/cmd/store"
)

type Router struct {
	*http.ServeMux
	uS *store.UrlStorage
}

func NewRouter(uS *store.UrlStorage) *Router {
	rt := &Router{
		ServeMux: http.NewServeMux(),
		uS:       uS,
	}

	rt.HandleFunc("/", http.HandlerFunc(rt.ComRequest).ServeHTTP)
	//rt.HandleFunc("/{id}", http.HandlerFunc(rt.ReturnOrignUrl).ServeHTTP)

	return rt
}

func (rt *Router) ComRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		defer r.Body.Close()

		content, err := io.ReadAll(r.Body)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		shrtUrl, err := rt.uS.CreateShortUrl(string(content))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shrtUrl))
	} else if r.Method == http.MethodGet {
		srtUrl := r.URL.Path
		srtUrl = srtUrl[1:]
		lngUrl, err := rt.uS.GetLongUrl(srtUrl)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Write([]byte(lngUrl))
	} else {
		http.Error(w, "bad method", http.StatusBadRequest)
		return
	}
}
