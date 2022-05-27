package main

import (
	"context"
	"github.com/botaevg/yandexgo/internal/handlers"
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
			h := http.HandlerFunc(handlers.PostHandler)
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

/*
func TestRouter(t *testing.T) {
	r := chi.NewRouter()
	ts := httptest.NewServer(r)
	ts.URL = "http://localhost:8080"
	ts.
	defer ts.Close()

	resp := testRequest(t, ts, "GET", "/testurl")
	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
	assert.Equal(t, "http://yandex.ru", resp.Header.Get("Location"))

	resp = testRequest(t, ts, "GET", "/tst")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, "", resp.Header.Get("Location"))

}

func testRequest(t *testing.T, ts *httptest.Server, method, path string) *http.Response {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	//fmt.Println(ts.URL + path)
	require.NoError(t, err)
	//http.DefaultClient.CheckRedirect
	http.DefaultClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp
}*/

func TestGetHandler(t *testing.T) {
	type want struct {
		code     int
		response string
		location string
	}
	tests := []struct {
		name    string
		want    want
		testURL string
	}{
		{
			name: "get test #1",
			want: want{
				code: http.StatusTemporaryRedirect,
				//response: "",
				location: "http://yandex.ru",
			},
			testURL: "testurl",
		},
		{
			name: "get test #2",
			want: want{
				code: http.StatusBadRequest,
				//response: "",
				location: "",
			},
			testURL: "Tst",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// теперь проверяем метод GET
			requestGet := httptest.NewRequest(http.MethodGet, "/{id}", nil)

			// создаём новый Recorder
			wGet := httptest.NewRecorder()
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.testURL)
			requestGet = requestGet.WithContext(context.WithValue(requestGet.Context(), chi.RouteCtxKey, rctx))
			// определяем хендлер
			hGet := http.HandlerFunc(handlers.GetHandler)
			// запускаем сервер
			hGet.ServeHTTP(wGet, requestGet)

			resGet := wGet.Result()
			defer resGet.Body.Close()
			assert.Equal(t, tt.want.code, resGet.StatusCode)

			assert.Equal(t, tt.want.location, resGet.Header.Get("Location"))
			/*resGetBody, err := ioutil.ReadAll(resGet.Body)
			require.NoError(t, err)
			err = resGet.Body.Close()
			require.NoError(t, err)

			//resGet.Location()

			assert.Equal(t, tt.want.location, string(resGetBody))
			*/
		})
	}
}
