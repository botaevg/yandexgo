package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io"
	"log"
	"math/rand"
	"net/http"
)

const letterAll = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type Subj struct {
	//ID        int    `json:"id"`
	URLorigin string `json:"urlorigin"`
	//Urlshort  string `json:"urlshort"`
}

var ListUrl = make(map[string]string)

func shortUrl() string {
	b := make([]byte, 5)
	for i := range b {
		b[i] = letterAll[rand.Intn(len(letterAll))]
	}
	return string(b)
}
func GetHandler(w http.ResponseWriter, r *http.Request) {
	// этот обработчик принимает только запросы, отправленные методом GET
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed!", http.StatusBadRequest)
		return
	}

	id := chi.URLParam(r, "id")

	w.Header().Set("Location", ListUrl[id])
	w.Header().Set("content-type", "application/json")

	w.WriteHeader(307)
	w.Write([]byte(ListUrl[id])) //

}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	// читаем Body
	if r.Method != http.MethodPost {
		http.Error(w, "Only Post requests are allowed!", http.StatusBadRequest)
		return
	}

	//var subj Subj
	b, err := io.ReadAll(r.Body)
	// обрабатываем ошибку
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	} /*else {
		err = json.Unmarshal(b, &subj)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}*/
	shortURL := shortUrl() // subj.URLorigin + "_short"
	ListUrl[shortURL] = string(b)
	w.WriteHeader(201)
	//fmt.Fprintln(w, shortURL)
	w.Write([]byte(shortURL))

	//w.Write([]byte(b))
}

func main() {
	r := chi.NewRouter()
	// зададим встроенные middleware, чтобы улучшить стабильность приложения
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/{id}", GetHandler)
	r.Post("/", PostHandler)

	// запуск сервера с адресом localhost, порт 8080
	log.Fatal(http.ListenAndServe(":8080", r))
}
