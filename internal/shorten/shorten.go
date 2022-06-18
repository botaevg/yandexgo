package shorten

import (
	"math/rand"
)

const (
	letterAll = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
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
