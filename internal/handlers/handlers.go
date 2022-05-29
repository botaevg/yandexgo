package handlers

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"io"
	"math/rand"
	"net/http"
)

const (
	letterAll = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	basicURL  = "http://localhost:8080/"
)

var ListURL = make(map[string]string)

func GetHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	if _, ok := ListURL[id]; !ok {
		http.Error(w, errors.New("BadRequest").Error(), http.StatusBadRequest)
		return
	}

	if ListURL[id] == "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Не найдено"))
		return
	}

	w.Header().Set("Location", ListURL[id])
	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Write([]byte(ListURL[id]))

}

func shortURL() string {
	b := make([]byte, 5)

	for i := range b {

		b[i] = letterAll[rand.Intn(len(letterAll))]
	}
	if _, ok := ListURL[string(b)]; ok {
		return shortURL()
	}
	return string(b)
}

func PostHandler(w http.ResponseWriter, r *http.Request) {

	b, err := io.ReadAll(r.Body)
	// обрабатываем ошибку
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if string(b) != "" {
		shortURLs := shortURL()
		ListURL[shortURLs] = string(b)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(basicURL + shortURLs))
	} else {
		err := errors.New("BadRequest")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

}
