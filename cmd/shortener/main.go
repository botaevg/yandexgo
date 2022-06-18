package main

import (
	"flag"
	"github.com/botaevg/yandexgo/internal/config"
	"github.com/botaevg/yandexgo/internal/handlers"
	"github.com/botaevg/yandexgo/internal/repositories"

	"github.com/botaevg/yandexgo/internal/middleapp"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	appConfig := config.GetConfig()
	flag.Parse()
	myApp := NewApp(appConfig)
	myApp.Run()

}

type App struct {
	config config.Config
}

func NewApp(cfg config.Config) App {
	return App{
		config: cfg,
	}
}

func (a App) Run() {
	r := chi.NewRouter()
	// зададим встроенные middleware, чтобы улучшить стабильность приложения
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleapp.GzipHandle)

	var storage repositories.Storage
	if a.config.FileStoragePath != "" {
		storage = repositories.FileStorage{
			a.config.FileStoragePath,
		}
	} else {
		storage = repositories.InMemoryStorage{}
	}
	h := handlers.New(a.config, storage)

	r.Post("/api/shorten", h.APIPost)
	r.Get("/{id}", h.GetHandler)
	r.Post("/", h.PostHandler)
	//r.Post("/", handlers.PostHandler)

	// запуск сервера с адресом localhost, порт 8080
	/*serverAddress, exists := os.LookupEnv("SERVER_ADDRESS")
	if exists {
		log.Print(serverAddress)
	}*/

	log.Fatal(http.ListenAndServe(a.config.ServerAddress, r))
}
