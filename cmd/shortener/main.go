package main

import (
	"flag"
	"github.com/botaevg/yandexgo/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

//const SERVER_ADDRESS = ":8080"

func init() {

	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
	//os.Setenv("SERVER_ADDRESS", )
	flag.String("a", ":8080", "server address")
	flag.String("b", "http://localhost:8080/", "base URL")
	flag.String("f", "shortlist.txt", "file storage path")
}
func main() {
	rand.Seed(time.Now().UnixNano())

	flag.Parse()
	r := chi.NewRouter()
	// зададим встроенные middleware, чтобы улучшить стабильность приложения
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(handlers.GzipHandle)
	r.Post("/api/shorten", handlers.APIPost)
	r.Get("/{id}", handlers.GetHandler)
	r.Post("/", handlers.PostHandler)

	// запуск сервера с адресом localhost, порт 8080
	serverAddress, exists := os.LookupEnv("SERVER_ADDRESS")
	if exists {
		log.Print(serverAddress)
	}

	log.Fatal(http.ListenAndServe(serverAddress, r))
}
