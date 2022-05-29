package main

import (
	"github.com/botaevg/yandexgo/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const port = ":8080"

func main() {
	rand.Seed(time.Now().UnixNano())

	r := chi.NewRouter()
	// зададим встроенные middleware, чтобы улучшить стабильность приложения
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/{id}", handlers.GetHandler)
	r.Post("/", handlers.PostHandler)

	// запуск сервера с адресом localhost, порт 8080
	log.Fatal(http.ListenAndServe(port, r))
}
