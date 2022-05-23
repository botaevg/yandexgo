package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPostHandler(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name      string
		want      want
		inputBody string
	}{
		// TODO: Add test cases.
		{
			name: "post test #1",
			want: want{
				code:        201,
				response:    "short",
				contentType: "",
			},
			inputBody: `{"urlorigin": "http://www.example.com"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.inputBody))
			request.Header.Set("content-type", "application/json")
			w := httptest.NewRecorder()
			h := http.HandlerFunc(PostHandler)
			h.ServeHTTP(w, request)
			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)

			resBody, err := ioutil.ReadAll(res.Body)
			require.NoError(t, err)
			err = res.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.response, string(resBody))
		})
	}
}

func TestGetHandler(t *testing.T) {
	type want struct {
		code     int
		response string
		location string
	}
	tests := []struct {
		name      string
		want      want
		inputBody string
	}{
		{
			name: "get test #2",
			want: want{
				code:     307,
				response: "",
				location: "yandex",
			},
			inputBody: `{"urlorigin": "http://www.example.com"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//записываем тестовые данные
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.inputBody))
			request.Header.Set("content-type", "application/json")
			w := httptest.NewRecorder()
			h := http.HandlerFunc(PostHandler)
			h.ServeHTTP(w, request)
			//res := w.Result()
			//fmt.Println(res)

			// теперь проверяем метод GET
			requestGet := httptest.NewRequest(http.MethodGet, "/short", nil)

			// создаём новый Recorder
			wGet := httptest.NewRecorder()
			// определяем хендлер
			hGet := http.HandlerFunc(GetHandler)
			// запускаем сервер
			hGet.ServeHTTP(wGet, requestGet)
			resGet := wGet.Result()
			assert.Equal(t, tt.want.code, resGet.StatusCode)

			assert.Equal(t, tt.want.location, resGet.Header.Get("Location"))
		})
	}
}
