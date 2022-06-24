package shorten

import (
	"math/rand"
)

const (
	letterAll = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digitAll  = "0123456789"
)

func ShortURL() string {
	b := make([]byte, 5)

	for i := range b {

		b[i] = letterAll[rand.Intn(len(letterAll))]
	}
	/*if _, ok := handlers.ListURL[string(b)]; ok {
		return ShortURL()
	}*/
	return string(b)
}

func GeneratorID() string {
	b := make([]byte, 5)

	for i := range b {

		b[i] = letterAll[rand.Intn(len(digitAll))]
	}
	/*if _, ok := handlers.ListURL[string(b)]; ok {
		return ShortURL()
	}*/
	return string(b)
}
