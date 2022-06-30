package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`    // envDefault:":8080"`
	BaseURL         string `env:"BASE_URL"`          // envDefault:"http://localhost:8080/"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"` // envDefault:"shortlist.txt"`
	DATABASEDSN     string `env:"DATABASE_DSN"`
}

func GetConfig() (Config, error) {

	cfg := Config{}
	flag.StringVar(&cfg.ServerAddress, "a", ":8080", "port to listen on")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080/", "base url")
	flag.StringVar(&cfg.FileStoragePath, "f", "shortlist.txt", "file storage path")
	flag.StringVar(&cfg.DATABASEDSN, "d", "postgresql://postgres:sqllife@localhost:5434/yandexgo", "data base dns")
	//postgresql://postgres:sqllife@localhost:5434/yandexgo
	//FILE_STORAGE_PATH=shortlist.txt
	//DATABASE_DSN=postgresql://postgres:sqllife@localhost:5434/yandexgo
	flag.Parse()
	if err := env.Parse(&cfg); err != nil {
		//fmt.Printf("%+v\n", err)
		return Config{}, err
	}
	if len(cfg.BaseURL) > 0 {
		x := cfg.BaseURL[len(cfg.BaseURL)-1]
		if string(x) != "/" {
			cfg.BaseURL += "/"
		}
	}

	return cfg, nil
}
