package middleapp

import (
	"compress/gzip"
	"context"
	"github.com/botaevg/yandexgo/internal/cookies"
	"github.com/botaevg/yandexgo/internal/repositories"
	"io"
	"log"
	"net/http"
	"strings"
)

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
			http.Error(w, err.Error(), http.StatusInternalServerError)
			//io.WriteString(w, err.Error())
			return
		}
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")

		if r.Header.Get(`Content-Encoding`) == `gzip` {
			gzb, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			r.Body = gzb
			defer gz.Close()
		}

		// передаём обработчику страницы переменную типа gzipWriter для вывода данных
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

type AuthKey string

type AuthMiddleware struct {
	storage repositories.Storage
}

func NewAuthMiddleware(storage repositories.Storage) *AuthMiddleware {
	return &AuthMiddleware{
		storage: storage,
	}
}

func (a AuthMiddleware) CheckCookie(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idUser := cookies.VerificationCookie(a.storage, r, &w)

		/*
			x, err := r.Cookie("id")
			if err != nil {
				log.Print("нет такого кука")
				value := ""
				http.SetCookie(w, &http.Cookie{
					Name:  "id",
					Value: value,
				})
			} else {

				log.Print(x)
			}*/
		log.Print("testtt")
		log.Print(idUser)
		log.Print("testtt")
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), AuthKey("idUser"), idUser)))

	})
}
