package repositories

import (
	"encoding/hex"
	"errors"
	"log"
	"os"
	"strings"
)

type Storage interface {
	AddShort(string, string, string) error
	GetFullURL(string) (string, error)
	AddCookie(string, []byte, []byte) error
	GetID(string) ([][]byte, error)
	GetAllShort(string) ([]URLpair, error)
}

type FileStorage struct {
	FileStorage string
}

type URLpair struct {
	ShortURL string `json:"ShortURL"`
	FullURL  string `json:"OriginalURL"`
}

type URLUser struct {
	AllURL []URLpair
}

type InMemoryStorage struct {
	dataURL    map[string][]string
	dataCookie map[string][][]byte
}

func (f InMemoryStorage) GetAllShort(idUser string) ([]URLpair, error) {
	var urlUser []URLpair
	for key, value := range f.dataURL {
		if value[1] == idUser {
			x := URLpair{
				FullURL:  value[0],
				ShortURL: "http://localhost:8080/" + key,
			}
			urlUser = append(urlUser, x)
		}
	}
	if len(urlUser) == 0 {
		return urlUser, errors.New("нет URL пользователя")
	}
	return urlUser, nil
}

func (f FileStorage) GetAllShort(idUser string) ([]URLpair, error) {
	var urlUser []URLpair

	data, err := os.ReadFile(f.FileStorage)
	if err != nil {
		return urlUser, errors.New("неоткрылся файл")
	}

	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, idUser) {
			short := strings.Split(line, ":")[1]
			full := strings.Join(strings.Split(line, ":")[2:], ":")
			x := URLpair{
				FullURL:  full,
				ShortURL: "http://localhost:8080/" + short,
			}
			urlUser = append(urlUser, x)
		}
	}

	if len(urlUser) == 0 {
		return urlUser, errors.New("нет URL пользователя")
	}
	return urlUser, nil
}

func (f FileStorage) GetFullURL(id string) (string, error) {

	data, err := os.ReadFile(f.FileStorage)
	if err != nil {
		return "", err
	}

	for _, line := range strings.Split(string(data), "\n") {
		if strings.Contains(line, id) {
			return strings.Join(strings.Split(line, ":")[2:], ":"), nil

		}
	}
	return "", errors.New("BadRequest")

}

func (f InMemoryStorage) GetFullURL(id string) (string, error) {

	if _, ok := f.dataURL[id]; !ok {

		return "", errors.New("BadRequest")
	}

	return f.dataURL[id][0], nil
}

func (f FileStorage) AddShort(body string, s string, idUser string) error {
	file, err := os.OpenFile(f.FileStorage, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)

	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(idUser + ":" + s + ":" + body + "\n")
	if err != nil {
		return err
	}
	return nil
}

func (f InMemoryStorage) AddShort(body string, s string, idUser string) error {
	f.dataURL[s] = []string{body, idUser}
	return nil
}

func (f FileStorage) AddCookie(idEncrypt string, key []byte, nonce []byte) error {
	file, err := os.OpenFile(f.FileStorage, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)

	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(idEncrypt + ":::" + hex.EncodeToString(key) + ":::" + hex.EncodeToString(nonce) + "\n")
	if err != nil {
		return err
	}
	return nil
}

func (f InMemoryStorage) AddCookie(idEncrypt string, key []byte, nonce []byte) error {
	f.dataCookie[idEncrypt] = [][]byte{key, nonce}
	if len(f.dataCookie) == 0 {
		log.Print("запись не удалась")
	}
	return nil
}

func (f FileStorage) GetID(idEncrypt string) ([][]byte, error) {
	data, err := os.ReadFile(f.FileStorage)
	if err != nil {
		return [][]byte{}, err
	}

	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, idEncrypt) {
			key, nonce := strings.Split(line, ":::")[1], strings.Split(line, ":::")[2]
			k, _ := hex.DecodeString(key)
			n, _ := hex.DecodeString(nonce)
			return [][]byte{
				k,
				n,
			}, nil

		}
	}
	return [][]byte{}, errors.New("NO found cookie")
}

func (f InMemoryStorage) GetID(idEncrypt string) ([][]byte, error) {
	log.Print("пытаемся получить ключ и вектор: " + idEncrypt)
	if _, ok := f.dataCookie[idEncrypt]; !ok {
		return [][]byte{}, errors.New("cookie not found")
	}
	return f.dataCookie[idEncrypt], nil
}

func NewFileStorage(p string) *FileStorage {
	return &FileStorage{
		FileStorage: p,
	}
}

func NewInMemoryStorage() *InMemoryStorage {
	IMS := InMemoryStorage{}
	IMS.dataURL = make(map[string][]string)
	IMS.dataCookie = make(map[string][][]byte)

	return &IMS
}
