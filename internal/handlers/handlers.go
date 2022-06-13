package handlers

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
)

const (
	letterAll = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	//BASE_URL          = "http://localhost:8080/"
	//FILE_STORAGE_PATH = "shortlist.txt"
)

type URL struct {
	FullURL  string `json:"url"`
	ShortURL string `json:"result"`
}

var ListURL = make(map[string]string)

func GetHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	var u string
	fileStorage, exists := os.LookupEnv("FILE_STORAGE_PATH")
	if !exists || fileStorage == "" {

		if _, ok := ListURL[id]; !ok {
			http.Error(w, errors.New("BadRequest").Error(), http.StatusBadRequest)
			return
		}
		u = ListURL[id]

	} else {
		data, err := os.ReadFile(fileStorage)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		foundURL := false
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, id) {
				u = strings.Join(strings.Split(line, ":")[1:], ":")
				log.Print(u)
				foundURL = true
				break
			}
		}
		if !foundURL {
			http.Error(w, errors.New("BadRequest").Error(), http.StatusBadRequest)
			return
		}
	}
	if u == "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Не найдено"))
		return
	}

	w.Header().Set("Location", u)
	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Write([]byte(u))

}

func shortURL() string {
	b := make([]byte, 5)

	for i := range b {

		b[i] = letterAll[rand.Intn(len(letterAll))]
	}
	if _, ok := ListURL[string(b)]; ok {
		return shortURL()
	}
	return string(b)
}

func PostHandler(w http.ResponseWriter, r *http.Request) {

	PostFull(w, r, false)

}

func APIPost(w http.ResponseWriter, r *http.Request) {
	PostFull(w, r, true)
	/*
		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		obj := URL{}

		if err := json.Unmarshal(b, &obj); err != nil {
			panic(err)
		}
		if obj.FullURL != "" {
			shortURLs := shortURL()
			ListURL[shortURLs] = obj.FullURL
			w.WriteHeader(http.StatusCreated)
			w.Header().Set("content-type", "application/json")
			baseURL, exists := os.LookupEnv("BASE_URL")
			if exists {
				log.Print(baseURL)
			}
			w.Write([]byte(`{"result":"` + baseURL + shortURLs + `"}`))
		} else {
			err := errors.New("BadRequest")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}*/

}

func PostFull(w http.ResponseWriter, r *http.Request, isAPI bool) {

	var reader io.Reader

	if r.Header.Get(`Content-Encoding`) == `gzip` {
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		reader = gz
		defer gz.Close()
	} else {
		reader = r.Body
	}

	b, err := io.ReadAll(reader)
	// обрабатываем ошибку
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//var strURL string
	strURL := string(b)
	if isAPI {
		obj := URL{}
		if err := json.Unmarshal(b, &obj); err != nil {
			panic(err)
		}
		strURL = obj.FullURL
	}

	if strURL != "" {
		shortURLs := shortURL()
		fileStorage, exists := os.LookupEnv("FILE_STORAGE_PATH")
		if !exists || fileStorage == "" {
			ListURL[shortURLs] = string(b)
		} else {
			file, err := os.OpenFile(fileStorage, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer file.Close()
			file.WriteString(shortURLs + ":" + strURL + "\n")
		}

		baseURL, _ := os.LookupEnv("BASE_URL")
		if isAPI {
			w.Header().Set("content-type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"result":"` + baseURL + shortURLs + `"}`))
		} else {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(baseURL + shortURLs))
		}
	} else {
		err := errors.New("BadRequest")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	// w.Writer будет отвечать за gzip-сжатие, поэтому пишем в него
	return w.Writer.Write(b)
}
func GzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// проверяем, что клиент поддерживает gzip-сжатие
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			// если gzip не поддерживается, передаём управление
			// дальше без изменений
			next.ServeHTTP(w, r)
			return
		}

		// создаём gzip.Writer поверх текущего w
		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		// передаём обработчику страницы переменную типа gzipWriter для вывода данных
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}
