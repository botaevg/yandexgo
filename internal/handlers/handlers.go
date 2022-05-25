package handlers

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"io"
	"math/rand"
	"net/http"
)

const letterAll = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var ListUrl = make(map[string]string)

func GetHandler(w http.ResponseWriter, r *http.Request) {
	// этот обработчик принимает только запросы, отправленные методом GET
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed!", http.StatusBadRequest)
		return
	}

	id := chi.URLParam(r, "id")

	w.Header().Set("Location", ListUrl[id])

	w.WriteHeader(307)
	fmt.Println(ListUrl[id])
	if ListUrl[id] == "" {
		w.Write([]byte("Не найдено"))
	} else {
		w.Write([]byte(ListUrl[id]))
	}

}

func shortUrl() string {
	b := make([]byte, 5)
	for i := range b {
		b[i] = letterAll[rand.Intn(len(letterAll))]
	}
	return string(b)
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	// читаем Body
	if r.Method != http.MethodPost {
		http.Error(w, "Only Post requests are allowed!", http.StatusBadRequest)
		return
	}

	b, err := io.ReadAll(r.Body)
	// обрабатываем ошибку
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	shortURL := shortUrl()

	ListUrl[shortURL] = string(b)
	w.WriteHeader(201)
	w.Write([]byte("http://localhost:8080/" + shortURL))

}
