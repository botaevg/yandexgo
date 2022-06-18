package repositories

import (
	"errors"
	"os"
	"strings"
)

var ListURL = make(map[string]string)

type Storage interface {
	AddShort(string, string) error
	GetFullURL(string) (string, error)
}

type FileStorage struct {
	FileStorage string
}

type InMemoryStorage struct {
}

func (f FileStorage) GetFullURL(id string) (string, error) {
	var u string

	data, err := os.ReadFile(f.FileStorage)
	if err != nil {
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		return u, err
	}
	foundURL := false
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, id) {
			u = strings.Join(strings.Split(line, ":")[1:], ":")
			//log.Print(u)
			foundURL = true
			break
		}
	}
	if !foundURL {
		//http.Error(w, errors.New("BadRequest").Error(), http.StatusBadRequest)
		return u, errors.New("BadRequest")
	}
	return u, nil
}

func (f InMemoryStorage) GetFullURL(id string) (string, error) {
	var u string
	if _, ok := ListURL[id]; !ok {

		return u, errors.New("BadRequest")
	}
	u = ListURL[id]
	return u, nil
}
func (f FileStorage) AddShort(body string, s string) error {
	file, err := os.OpenFile(f.FileStorage, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)

	if err != nil {
		return err
	}
	defer file.Close()
	file.WriteString(s + ":" + body + "\n")
	return nil
}

func (f InMemoryStorage) AddShort(body string, s string) error {
	//strURL := string(s)
	ListURL[s] = body
	return nil
}
