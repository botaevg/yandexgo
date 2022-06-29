package main

import (
	"context"
	"github.com/botaevg/yandexgo/internal/config"
	"github.com/botaevg/yandexgo/internal/handlers"
	"github.com/botaevg/yandexgo/internal/repositories"

	"github.com/botaevg/yandexgo/internal/middleapp"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	appConfig, err := config.GetConfig()
	if err != nil {
		log.Print("Config error")
		return
	}
	//flag.Parse()

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
	r.Use(middleapp.CheckCookie)

	var storage repositories.Storage

	if a.config.DATABASEDSN != "" {
		log.Print("репо db")
		postgreSQLClient, err := repositories.NewClient(context.TODO(), 3, a.config.DATABASEDSN)
		if err != nil {
			log.Print("clien postgres fail")
		}
		storage = repositories.NewDB(postgreSQLClient)

	} else if a.config.FileStoragePath != "" {
		log.Print("репо файл")
		storage = repositories.NewFileStorage(a.config.FileStoragePath)
	} else {
		log.Print("репо мапа")
		storage = repositories.NewInMemoryStorage()
	}
	h := handlers.New(a.config, storage)

	r.Get("/api/user/urls", h.GetAllShortURL)
	r.Post("/api/shorten/batch", h.ApiShortBatch)
	r.Post("/api/shorten", h.APIPost)
	r.Get("/ping", h.CheckPing)
	r.Get("/{id}", h.GetHandler)
	r.Post("/", h.PostHandler)

	log.Fatal(http.ListenAndServe(a.config.ServerAddress, r))
}
