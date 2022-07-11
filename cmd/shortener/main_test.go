package main

import (
	"context"
	"fmt"
	"github.com/botaevg/yandexgo/internal/config"
	"github.com/botaevg/yandexgo/internal/handlers"
	"github.com/botaevg/yandexgo/internal/repositories"
	"github.com/go-chi/chi/v5"
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
				code:     http.StatusCreated,
				response: "",
				//contentType: "",
			},
			inputBody: "http://www.example.com",
		},
		{
			name: "post test #2",
			want: want{
				code:     http.StatusBadRequest,
				response: "",
				//contentType: "",
			},
			inputBody: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.inputBody))
			//request.Header.Set("content-type", "application/json")
			w := httptest.NewRecorder()
			asyncExecutionChannel := make(chan handlers.DeleteURL)
			x := handlers.New(config.Config{
				ServerAddress:   ":8080",
				BaseURL:         "http://localhost:8080/",
				FileStoragePath: "shortlist.txt",
			},
				repositories.FileStorage{
					FileStorage: "shortlist.txt",
				},
				asyncExecutionChannel,
			)
			h := http.HandlerFunc(x.PostHandler)
			h.ServeHTTP(w, request)
			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)

			resBody, err := ioutil.ReadAll(res.Body)
			require.NoError(t, err)
			err = res.Body.Close()
			require.NoError(t, err)

			assert.NotEqual(t, tt.want.response, string(resBody))
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
		testURL   string
		inputBody string
	}{
		{
			name: "get test #1",
			want: want{
				code: http.StatusTemporaryRedirect,
				//response: "",
				location: "http://www.example.com",
			},
			testURL:   "testurl",
			inputBody: "http://www.example.com",
		},
		{
			name: "get test #2",
			want: want{
				code: http.StatusBadRequest,
				//response: "",
				location: "",
			},
			testURL:   "Tst",
			inputBody: "http://www.example.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.inputBody))
			//request.Header.Set("content-type", "application/json")
			w := httptest.NewRecorder()
			asyncExecutionChannel := make(chan handlers.DeleteURL)
			x := handlers.New(config.Config{

				ServerAddress:   ":8080",
				BaseURL:         "http://localhost:8080/",
				FileStoragePath: "shortlist.txt",
			},
				repositories.FileStorage{
					FileStorage: "shortlist.txt",
				},
				asyncExecutionChannel,
			)
			h := http.HandlerFunc(x.PostHandler)
			h.ServeHTTP(w, request)
			res := w.Result()
			//assert.Equal(t, tt.want.code, res.StatusCode)

			resBody, err := ioutil.ReadAll(res.Body)
			require.NoError(t, err)
			err = res.Body.Close()
			require.NoError(t, err)

			//assert.NotEqual(t, tt.want.response, string(resBody))
			pathSlice := strings.Split(string(resBody), "/")
			pathLast := pathSlice[(len(pathSlice) - 1)]
			fmt.Println(string(pathLast))

			// теперь проверяем метод GET
			requestGet := httptest.NewRequest(http.MethodGet, "/{id}", nil)
			// создаём новый Recorder
			wGet := httptest.NewRecorder()
			rctx := chi.NewRouteContext()
			if tt.want.location == "" {
				rctx.URLParams.Add("id", tt.testURL)
			} else {
				rctx.URLParams.Add("id", pathLast)
			}

			requestGet = requestGet.WithContext(context.WithValue(requestGet.Context(), chi.RouteCtxKey, rctx))
			// определяем хендлер
			hGet := http.HandlerFunc(x.GetHandler)
			// запускаем сервер
			hGet.ServeHTTP(wGet, requestGet)

			resGet := wGet.Result()
			defer resGet.Body.Close()
			assert.Equal(t, tt.want.code, resGet.StatusCode)

			assert.Equal(t, tt.want.location, resGet.Header.Get("Location"))
			t.Log(resGet.Header.Get("Location"))
			t.Log(tt.want.location)

		})
	}
}

func TestApiPost(t *testing.T) {
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
		{
			name: "api post test #1",
			want: want{
				code:        http.StatusCreated,
				response:    "",
				contentType: "application/json",
			},
			inputBody: `{"url":"http://yandex.ru"}`,
		},
		{
			name: "api post test #2",
			want: want{
				code:        http.StatusBadRequest,
				response:    "",
				contentType: "text/plain; charset=utf-8",
			},
			inputBody: `{"url":"yandex"}`,
		},
		{
			name: "api post test #3",
			want: want{
				code:        http.StatusBadRequest,
				response:    "",
				contentType: "text/plain; charset=utf-8",
			},
			inputBody: "http://yandex.ru",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(tt.inputBody))
			request.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			asyncExecutionChannel := make(chan handlers.DeleteURL)
			x := handlers.New(config.Config{
				ServerAddress:   ":8080",
				BaseURL:         "http://localhost:8080/",
				FileStoragePath: "",
			},
				repositories.NewInMemoryStorage(),
				asyncExecutionChannel,
			)
			h := http.HandlerFunc(x.APIPost)
			h.ServeHTTP(w, request)
			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
			resBody, err := ioutil.ReadAll(res.Body)
			require.NoError(t, err)
			err = res.Body.Close()
			require.NoError(t, err)

			assert.NotEqual(t, tt.want.response, string(resBody))
		})
	}
}
