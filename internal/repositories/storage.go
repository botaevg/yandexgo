package repositories

import (
	"context"
	"encoding/hex"
	"errors"
	"github.com/botaevg/yandexgo/internal/domain"
	"github.com/botaevg/yandexgo/internal/shorten"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"os"
	"strings"
)

type Storage interface {
	AddShort(string, string, string) (string, error)
	GetFullURL(string) (string, error)
	AddCookie(string, []byte, []byte) error
	GetID(string) ([][]byte, error)
	GetAllShort(string) ([]URLpair, error)
	Ping(ctx context.Context) error
	//FindShort(string) (string, error)
	AddShortBatch([]domain.APIOriginBatch, string, string) ([]domain.APIShortBatch, error)
}

type FileStorage struct {
	FileStorage string
}

type URLpair struct {
	ShortURL string `json:"short_url"`
	FullURL  string `json:"original_url"`
}

type DBStorage struct {
	db *pgxpool.Pool
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
		log.Print(urlUser)
		return urlUser, errors.New("нет URL пользователя")
	}
	log.Print(urlUser)
	return urlUser, nil
}

func (f FileStorage) GetAllShort(idUser string) ([]URLpair, error) {
	var urlUser []URLpair

	data, err := os.ReadFile(f.FileStorage)
	if err != nil {
		log.Print(urlUser)
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
		log.Print(urlUser)
		return urlUser, errors.New("нет URL пользователя")
	}
	log.Print(urlUser)
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

func (f FileStorage) AddShort(body string, s string, idUser string) (string, error) {
	file, err := os.OpenFile(f.FileStorage, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)

	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = file.WriteString(idUser + ":" + s + ":" + body + "\n")
	if err != nil {
		return "", err
	}
	return s, nil
}

func (f InMemoryStorage) AddShort(body string, s string, idUser string) (string, error) {
	f.dataURL[s] = []string{body, idUser}
	return s, nil
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
			k, err := hex.DecodeString(key)
			if err != nil {
				return [][]byte{}, err
			}
			n, err := hex.DecodeString(nonce)
			if err != nil {
				return [][]byte{}, err
			}
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

func NewDB(pool *pgxpool.Pool) *DBStorage {
	return &DBStorage{
		db: pool,
	}
}

func (f DBStorage) AddShort(fullURL string, shortURL string, idEncrypt string) (string, error) {
	q := `
	INSERT INTO urls
(idEncrypt, shortURL, fullURL)
VALUES 
($1,$2,$3)
;
`
	//ON CONFLICT (fullURL) DO NOTHING
	_, err := f.db.Exec(context.Background(), q, idEncrypt, shortURL, fullURL)
	if err != nil {
		log.Print("Запись не создана")
		log.Print(err)
		//return errors.New("запись не добавлена")
		q := `
		SELECT shortURL FROM urls WHERE fullURL = $1
		`
		rows, err := f.db.Query(context.Background(), q, fullURL)
		if err != nil {
			return "", err
		}
		defer rows.Close()

		var short string
		for rows.Next() {
			err = rows.Scan(&short)
			if err != nil {
				return "", err
			}
		}

		// проверяем на ошибки
		err = rows.Err()
		if err != nil {
			return "", err
		}
		return short, nil
	} else {
		log.Print("Запись создана")
	}
	return shortURL, nil

}

func (f InMemoryStorage) AddShortBatch(origins []domain.APIOriginBatch, baseURL string, idEncrypt string) ([]domain.APIShortBatch, error) {
	var shortBatch []domain.APIShortBatch
	for _, v := range origins {
		shortURLs := shorten.ShortURL()
		f.dataURL[shortURLs] = []string{v.Origin, idEncrypt}
		shortBatch = append(shortBatch, domain.APIShortBatch{
			ID:       v.ID,
			ShortURL: baseURL + shortURLs,
		})
	}

	return shortBatch, nil
}

func (f FileStorage) AddShortBatch(origins []domain.APIOriginBatch, baseURL string, idEncrypt string) ([]domain.APIShortBatch, error) {
	file, err := os.OpenFile(f.FileStorage, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)

	if err != nil {
		return []domain.APIShortBatch{}, nil
	}
	defer file.Close()

	var shortBatch []domain.APIShortBatch

	for _, v := range origins {
		shortURLs := shorten.ShortURL()
		_, err = file.WriteString(idEncrypt + ":" + shortURLs + ":" + v.Origin + "\n")
		if err != nil {
			return []domain.APIShortBatch{}, nil
		}
		shortBatch = append(shortBatch, domain.APIShortBatch{
			ID:       v.ID,
			ShortURL: baseURL + shortURLs,
		})
	}

	return shortBatch, nil

}

func (f DBStorage) AddShortBatch(origins []domain.APIOriginBatch, baseURL string, idEncrypt string) ([]domain.APIShortBatch, error) {
	// шаг 1 — объявляем транзакцию
	tx, err := f.db.Begin(context.Background())
	if err != nil {
		return []domain.APIShortBatch{}, err
	}
	// шаг 1.1 — если возникает ошибка, откатываем изменения
	defer tx.Rollback(context.Background())

	q := `
	INSERT INTO urls
(idEncrypt, shortURL, fullURL)
VALUES 
($1,$2,$3)
;
`

	var shortBatch []domain.APIShortBatch

	for _, v := range origins {
		// шаг 3 — указываем, что каждое видео будет добавлено в транзакцию
		shortURLs := shorten.ShortURL()
		if _, err = tx.Exec(context.Background(), q, idEncrypt, shortURLs, v.Origin); err != nil {
			return []domain.APIShortBatch{}, err
		}
		shortBatch = append(shortBatch, domain.APIShortBatch{
			ID:       v.ID,
			ShortURL: baseURL + shortURLs,
		})
	}
	// шаг 4 — сохраняем изменения
	return shortBatch, tx.Commit(context.Background())
	///////////////

}

func (f DBStorage) GetFullURL(shortURL string) (string, error) {
	q := `
	SELECT idEncrypt, fullURL FROM urls WHERE shortURL = $1
`
	rows, err := f.db.Query(context.Background(), q, shortURL)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var id, fullURL string
	for rows.Next() {
		err = rows.Scan(&id, &fullURL)

		if err != nil {
			return "", err
		}
	}
	err = rows.Err()
	if err != nil {
		return "", err
	}

	return fullURL, nil
}

func (f DBStorage) AddCookie(idEncrypt string, key []byte, nonce []byte) error {
	q := `
	INSERT INTO cookies
(idEncrypt, key, nonce)
VALUES 
($1,$2,$3)
`
	_, err := f.db.Exec(context.Background(), q, idEncrypt, hex.EncodeToString(key), hex.EncodeToString(nonce))
	if err != nil {
		log.Print("Запись не создана")
		log.Print(err)
	} else {
		log.Print("Запись создана")
	}

	return nil
}
func (f DBStorage) GetID(idEncrypt string) ([][]byte, error) {
	q := `
	SELECT key, nonce FROM cookies WHERE idEncrypt = $1
`
	rows, err := f.db.Query(context.Background(), q, idEncrypt)
	if err != nil {
		return [][]byte{}, err
	}
	defer rows.Close()

	var key, nonce string
	for rows.Next() {
		err = rows.Scan(&key, &nonce)
		if err != nil {
			return nil, err
		}
	}

	// проверяем на ошибки
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	k, err := hex.DecodeString(key)
	if err != nil {
		return [][]byte{}, err
	}
	n, err := hex.DecodeString(nonce)
	if err != nil {
		return [][]byte{}, err
	}
	log.Print("ид найден")
	return [][]byte{
		k,
		n,
	}, nil
}
func (f DBStorage) GetAllShort(idEncrypt string) ([]URLpair, error) {
	q := `
	SELECT shortURL, fullURL FROM urls WHERE idEncrypt = $1
`
	rows, err := f.db.Query(context.Background(), q, idEncrypt)
	if err != nil {
		return []URLpair{}, err
	}
	defer rows.Close()

	var urlUser []URLpair

	for rows.Next() {
		x := URLpair{}
		err = rows.Scan(&x.ShortURL, &x.FullURL)
		if err != nil {
			return []URLpair{}, err
		}

		urlUser = append(urlUser, x)

	}
	err = rows.Err()
	if err != nil {
		return []URLpair{}, err
	}

	return urlUser, nil

}

func (f DBStorage) Ping(ctx context.Context) error {
	if err := f.db.Ping(ctx); err != nil {
		log.Print("ping error")
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

func (f InMemoryStorage) Ping(ctx context.Context) error {

	return errors.New("repo map")
}

func (f FileStorage) Ping(ctx context.Context) error {

	return errors.New("repo file")
}

/*
func (f InMemoryStorage) FindShort(s string) (string, error) {
	return "", errors.New("repo map")
}

func (f FileStorage) FindShort(s string) (string, error) {
	return "", errors.New("repo file")
}

func (f DBStorage) FindShort(s string) (string, error) {
	q := `
	SELECT shortURL FROM urls WHERE fullURL = $1
`
	rows, err := f.db.Query(context.Background(), q, s)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var short string
	for rows.Next() {
		err = rows.Scan(&short)
		if err != nil {
			return "", err
		}
	}

	// проверяем на ошибки
	err = rows.Err()
	if err != nil {
		return "", err
	}
	return short, nil
}
*/
