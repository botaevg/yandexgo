package handlers

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"github.com/botaevg/yandexgo/internal/config"
	"github.com/botaevg/yandexgo/internal/repositories"
	"github.com/go-chi/chi/v5"
	"io"
	"log"

	"github.com/botaevg/yandexgo/internal/shorten"
	"net/http"
)

type handler struct {
	config  config.Config
	storage repositories.Storage
}

func New(cfg config.Config, storage repositories.Storage) *handler {
	return &handler{
		config:  cfg,
		storage: storage,
	}
}

type URL struct {
	FullURL  string `json:"url"`
	ShortURL string `json:"result"`
}

func (h *handler) GetHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	u, err := h.storage.GetFullURL(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if u == "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Не найдено"))
		return
	}

	w.Header().Set("Location", u)
	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Write([]byte(u))

}

func (h *handler) PostHandler(w http.ResponseWriter, r *http.Request) {

	var reader io.Reader

	if r.Header.Get(`Content-Encoding`) == `gzip` {
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		reader = gz
		defer gz.Close()
	} else {
		reader = r.Body
	}

	b, err := io.ReadAll(reader)
	// обрабатываем ошибку
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	strURL := string(b)

	if strURL == "" {
		http.Error(w, errors.New("BadRequest").Error(), http.StatusBadRequest)
		return
	}

	shortURLs := shorten.ShortURL()
	err = h.storage.AddShort(strURL, shortURLs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	baseURL := h.config.BaseURL // os.LookupEnv("BASE_URL")

	/*if len(baseURL) > 0 {
		x := baseURL[len(baseURL)-1]
		log.Print(x, string(x))
		if string(x) != "/" {
			baseURL += "/"
		}
	}*/
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(baseURL + shortURLs))

}

func (h *handler) APIPost(w http.ResponseWriter, r *http.Request) {

	var reader io.Reader

	if r.Header.Get(`Content-Encoding`) == `gzip` {
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		reader = gz
		defer gz.Close()
	} else {
		reader = r.Body
	}

	b, err := io.ReadAll(reader)
	// обрабатываем ошибку
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	obj := URL{}
	if err := json.Unmarshal(b, &obj); err != nil {
		http.Error(w, errors.New("BadRequest").Error(), http.StatusBadRequest)
		return
	}
	strURL := obj.FullURL

	if strURL == "" {
		http.Error(w, errors.New("BadRequest").Error(), http.StatusBadRequest)
		return
	}

	shortURLs := shorten.ShortURL()
	err = h.storage.AddShort(strURL, shortURLs)
	if err != nil {

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	baseURL := h.config.BaseURL // os.LookupEnv("BASE_URL")

	if len(baseURL) > 0 {
		x := baseURL[len(baseURL)-1]
		log.Print(x, string(x))
		if string(x) != "/" {
			baseURL += "/"
		}
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"result":"` + baseURL + shortURLs + `"}`))
}
