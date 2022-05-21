package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Subj struct {
	//ID        int    `json:"id"`
	URLorigin string `json:"urlorigin"`
	//Urlshort  string `json:"urlshort"`
}

var ListUrl = make(map[string]string)

var form = `<html>
    <head>
    <title></title>
    </head>
    <body>
        <form action="/login" method="post">
            <label>Логин</label><input type="text" name="login">
            <label>Пароль<input type="password" name="password">
            <input type="submit" value="Login">
        </form>
    </body>
</html>`

func GetHandler(w http.ResponseWriter, r *http.Request) {
	// этот обработчик принимает только запросы, отправленные методом GET
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
	// продолжаем обработку запроса
	// ...
	id := r.FormValue("id")
	w.WriteHeader(307)

	w.Write([]byte(id + "_short"))
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	// читаем Body
	if r.Method == http.MethodPost {
		//http.Error(w, "Only Post requests are allowed!", http.StatusMethodNotAllowed)
		//return
		var subj Subj
		b, err := io.ReadAll(r.Body)
		// обрабатываем ошибку
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		} else {
			err = json.Unmarshal(b, &subj)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
		}
		shortURL := subj.URLorigin + "_short"
		ListUrl[shortURL] = subj.URLorigin
		w.WriteHeader(201)
		fmt.Fprintln(w, subj.URLorigin)
		//w.Write([]byte(subj.URLorigin))

	} else if r.Method == http.MethodGet {
		//id := r.FormValue("id")
		//id := r.URL.Query().Get("id")
		path := []rune(r.URL.Path)
		id := string(path[1:])
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(307)
		//ListUrl[id] = id + "Short" // не тут надо записывать, а POST
		w.Header().Set("Location", ListUrl[id])
		w.Write([]byte(ListUrl[id])) //
	} else {
		http.Error(w, "Post and Get requests are allowed!", http.StatusBadRequest)
		return
	}
	//w.Write([]byte(b))
}

func main() {

	// маршрутизация запросов обработчику
	//http.HandleFunc("/:id", GetHandler)
	//mux := http.NewServeMux()
	//mux.HandleFunc("/", PostHandler)
	http.HandleFunc("/", PostHandler)

	// запуск сервера с адресом localhost, порт 8080
	server := &http.Server{
		Addr: ":8080",
	}
	server.ListenAndServe()
}
