package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/botaevg/yandexgo/internal/config"
	"github.com/botaevg/yandexgo/internal/cookies"
	"github.com/botaevg/yandexgo/internal/repositories"
	"github.com/botaevg/yandexgo/internal/shorten"
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"net/url"
	"time"

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

func (h *handler) CheckPing(w http.ResponseWriter, r *http.Request) {
	log.Print(h.config.DATABASEDSN)
	db, err := sql.Open("pgx", h.config.DATABASEDSN)
	if err != nil {
		log.Print("DB error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()
	// можем продиагностировать соединение

	_, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err = db.Ping(); err != nil {
		log.Print("ping error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("oki"))
}
func (h *handler) GetAllShortURL(w http.ResponseWriter, r *http.Request) {
	idUser := cookies.VerificationCookie(h.storage, r, &w)
	log.Print(idUser)

	var u []repositories.URLpair

	u, err := h.storage.GetAllShort(idUser)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte("Не найдено"))
		return
	}
	b, err := json.Marshal(u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	log.Print(b)
	log.Print(string(b))
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func (h *handler) GetHandler(w http.ResponseWriter, r *http.Request) {
	idUser := cookies.VerificationCookie(h.storage, r, &w)
	log.Print(idUser)

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
	idUser := cookies.VerificationCookie(h.storage, r, &w)
	log.Print(idUser)

	b, err := io.ReadAll(r.Body) //reader
	// обрабатываем ошибку
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	strURL := string(b)
	_, err = url.ParseRequestURI(strURL)
	if err != nil {
		http.Error(w, errors.New("BadRequest").Error(), http.StatusBadRequest)
		return
	}

	shortURLs := shorten.ShortURL()
	err = h.storage.AddShort(strURL, shortURLs, idUser)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	baseURL := h.config.BaseURL // os.LookupEnv("BASE_URL")

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(baseURL + shortURLs))

}

func (h *handler) APIPost(w http.ResponseWriter, r *http.Request) {
	idUser := cookies.VerificationCookie(h.storage, r, &w)
	log.Print(idUser)

	b, err := io.ReadAll(r.Body) //reader
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

	_, err = url.ParseRequestURI(strURL)
	if err != nil {
		http.Error(w, errors.New("BadRequest").Error(), http.StatusBadRequest)
		return
	}

	shortURLs := shorten.ShortURL()
	err = h.storage.AddShort(strURL, shortURLs, idUser)

	if err != nil {

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	baseURL := h.config.BaseURL // os.LookupEnv("BASE_URL")

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"result":"` + baseURL + shortURLs + `"}`))
}
