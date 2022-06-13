package config

type Config struct {
	SERVER_ADDRESS string `env:"SERVER_ADDRESS" envDefault:":8080"`
	BASE_URL       string `env:"BASE_URL" envDefault:"http://localhost:8080/"`
}

var Cfg Config

func GetConfig() Config {
	return Cfg
}
