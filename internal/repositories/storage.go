package repositories

import (
	"errors"
	"os"
	"strings"
)

type Storage interface {
	AddShort(string, string) error
	GetFullURL(string) (string, error)
}

type FileStorage struct {
	FileStorage string
}

type InMemoryStorage map[string]string

func (f FileStorage) GetFullURL(id string) (string, error) {

	data, err := os.ReadFile(f.FileStorage)
	if err != nil {
		return "", err
	}

	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, id) {
			return strings.Join(strings.Split(line, ":")[1:], ":"), nil

		}
	}
	return "", errors.New("BadRequest")

}

func (f InMemoryStorage) GetFullURL(id string) (string, error) {

	if _, ok := f[id]; !ok {

		return "", errors.New("BadRequest")
	}

	return f[id], nil
}

func (f FileStorage) AddShort(body string, s string) error {
	file, err := os.OpenFile(f.FileStorage, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)

	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(s + ":" + body + "\n")
	if err != nil {
		return err
	}
	return nil
}

func (f InMemoryStorage) AddShort(body string, s string) error {
	f[s] = body
	return nil
}

func NewFileStorage(p string) *FileStorage {
	return &FileStorage{
		FileStorage: p,
	}
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{}
}
