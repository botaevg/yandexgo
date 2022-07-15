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
	AddShort(domain.URLForAddStorage) (string, error)
	GetFullURL(string) (domain.URLForAddStorage, error)
	AddCookie(string, []byte, []byte) error
	GetID(string) ([][]byte, error)
	GetAllShort(string) ([]domain.URLForAddStorage, error)
	Ping(ctx context.Context) error
	AddShortBatch([]domain.URLForAddStorage) error
	UpdateFlagDelete([]string, string) error
}

type FileStorage struct {
	FileStorage string
}

type DBStorage struct {
	db *pgxpool.Pool
}

type InMemoryStorage struct {
	dataURL    map[string][]string
	dataCookie map[string][][]byte
}

func (f InMemoryStorage) UpdateFlagDelete(shorts []string, idUser string) error {
	for _, v := range shorts {
		if f.dataURL[v][1] == idUser {
			f.dataURL[v][2] = "true"
		}
	}
	return nil
}

func (f FileStorage) UpdateFlagDelete(shorts []string, idUser string) error {
	data, err := os.ReadFile(f.FileStorage)
	if err != nil {
		return err
	}
	newData, err := os.Create(f.FileStorage)
	if err != nil {
		return err
	}
	for _, line := range strings.Split(string(data), "\n") {

		lineSlice := strings.Split(line, ":")
		contain := false
		if len(lineSlice) > 2 {
			for _, v := range shorts {
				if v == lineSlice[1] {
					contain = true
					log.Print(contain)
				}
			}
		}
		if contain {

			_, err = newData.WriteString(lineSlice[0] + ":" + lineSlice[1] + ":" + "true" + ":" + strings.Join(lineSlice[3:], ":") + "\n")
			if err != nil {
				return err
			}
		} else {
			_, err = newData.WriteString(line + "\n")
			log.Print("запись")
			log.Print(line)
			if err != nil {
				return err
			}
		}
	}

	defer newData.Close()

	return nil
}

func (f DBStorage) UpdateFlagDelete(shorts []string, idUser string) error {

	/*params := make([]interface{}, len(shorts))
	for i, v := range shorts {
		params[i] = v
	}*/

	q := `
	update urls 
	set deleted = true 
	where shortURL = any($1) and idEncrypt = $2
`

	t, err := f.db.Exec(context.Background(), q, shorts, idUser)
	log.Print(t.RowsAffected())
	if err != nil {
		log.Print("Запись не обновлена")
		log.Print(err)
	}
	return nil
}

func (f InMemoryStorage) GetAllShort(idUser string) ([]domain.URLForAddStorage, error) {
	var urlUser []domain.URLForAddStorage
	for key, value := range f.dataURL {
		if value[1] == idUser {
			x := domain.URLForAddStorage{
				FullURL:  value[0],
				ShortURL: key,
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

func (f FileStorage) GetAllShort(idUser string) ([]domain.URLForAddStorage, error) {
	var urlUser []domain.URLForAddStorage

	data, err := os.ReadFile(f.FileStorage)
	if err != nil {
		log.Print(urlUser)
		return urlUser, errors.New("неоткрылся файл")
	}

	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, idUser) {
			short := strings.Split(line, ":")[1]
			full := strings.Join(strings.Split(line, ":")[3:], ":")
			x := domain.URLForAddStorage{
				FullURL:  full,
				ShortURL: short,
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

func (f DBStorage) GetAllShort(idEncrypt string) ([]domain.URLForAddStorage, error) {
	q := `
	SELECT shortURL, fullURL FROM urls WHERE idEncrypt = $1
`
	rows, err := f.db.Query(context.Background(), q, idEncrypt)
	if err != nil {
		return []domain.URLForAddStorage{}, err
	}
	defer rows.Close()

	var urlUser []domain.URLForAddStorage

	for rows.Next() {
		x := domain.URLForAddStorage{}
		err = rows.Scan(&x.ShortURL, &x.FullURL)
		if err != nil {
			return []domain.URLForAddStorage{}, err
		}

		urlUser = append(urlUser, x)

	}
	err = rows.Err()
	if err != nil {
		return []domain.URLForAddStorage{}, err
	}

	return urlUser, nil

}

func (f FileStorage) GetFullURL(id string) (domain.URLForAddStorage, error) {

	data, err := os.ReadFile(f.FileStorage)
	if err != nil {
		return domain.URLForAddStorage{}, err
	}

	for _, line := range strings.Split(string(data), "\n") {
		if strings.Contains(line, id) {
			var deleted bool
			if strings.Split(line, ":")[2] == "true" {
				deleted = true
			}
			return domain.URLForAddStorage{
				FullURL: strings.Join(strings.Split(line, ":")[3:], ":"),
				Deleted: deleted,
			}, nil

		}
	}
	return domain.URLForAddStorage{}, errors.New("BadRequest")

}

func (f InMemoryStorage) GetFullURL(id string) (domain.URLForAddStorage, error) {

	if _, ok := f.dataURL[id]; !ok {

		return domain.URLForAddStorage{}, errors.New("BadRequest")
	}

	var deleted bool
	if f.dataURL[id][2] == "true" {
		deleted = true
	}

	return domain.URLForAddStorage{
		FullURL: f.dataURL[id][0],
		Deleted: deleted,
	}, nil
}

func (f FileStorage) AddShort(item domain.URLForAddStorage) (string, error) {
	file, err := os.OpenFile(f.FileStorage, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)

	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = file.WriteString(item.IDUser + ":" + item.ShortURL + ":" + "false" + ":" + item.FullURL + "\n")
	if err != nil {
		return "", err
	}
	return item.ShortURL, nil
}

func (f InMemoryStorage) AddShort(item domain.URLForAddStorage) (string, error) {
	f.dataURL[item.ShortURL] = []string{item.FullURL, item.IDUser, "false"}
	return item.ShortURL, nil
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

func (f DBStorage) AddShort(item domain.URLForAddStorage) (string, error) {
	q := `
	INSERT INTO urls
(idEncrypt, shortURL, fullURL, deleted)
VALUES 
($1,$2,$3, false)
;
`
	//ON CONFLICT (fullURL) DO NOTHING
	_, err := f.db.Exec(context.Background(), q, item.IDUser, item.ShortURL, item.FullURL)
	if err != nil {
		log.Print("Запись не создана")
		log.Print(err)
		//return errors.New("запись не добавлена")
		q := `
		SELECT shortURL FROM urls WHERE fullURL = $1
		`
		rows, err := f.db.Query(context.Background(), q, item.FullURL)
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
	return item.ShortURL, nil

}

func (f InMemoryStorage) AddShortBatch(origins []domain.URLForAddStorage) error {
	for _, v := range origins {
		shortURLs := shorten.ShortURL()
		f.dataURL[shortURLs] = []string{v.FullURL, v.IDUser}

	}

	return nil
}

func (f FileStorage) AddShortBatch(origins []domain.URLForAddStorage) error {
	file, err := os.OpenFile(f.FileStorage, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)

	if err != nil {
		return nil
	}
	defer file.Close()

	for _, v := range origins {
		_, err = file.WriteString(v.IDUser + ":" + v.ShortURL + ":" + v.FullURL + "\n")
		if err != nil {
			return nil
		}

	}

	return nil

}

func (f DBStorage) AddShortBatch(origins []domain.URLForAddStorage) error {
	// шаг 1 — объявляем транзакцию
	tx, err := f.db.Begin(context.Background())
	if err != nil {
		return err
	}
	// шаг 1.1 — если возникает ошибка, откатываем изменения
	defer tx.Rollback(context.Background())

	q := `
	INSERT INTO urls
(idEncrypt, shortURL, fullURL, deleted)
VALUES 
($1,$2,$3, false)
;
`

	for _, v := range origins {
		// шаг 3 — указываем, что каждое видео будет добавлено в транзакцию
		if _, err = tx.Exec(context.Background(), q, v.IDUser, v.ShortURL, v.FullURL); err != nil {
			return err
		}

	}
	// шаг 4 — сохраняем изменения
	return tx.Commit(context.Background())
	///////////////

}

func (f DBStorage) GetFullURL(shortURL string) (domain.URLForAddStorage, error) {
	q := `
	SELECT idEncrypt, fullURL, deleted FROM urls WHERE shortURL = $1
`
	rows, err := f.db.Query(context.Background(), q, shortURL)
	if err != nil {
		return domain.URLForAddStorage{}, err
	}
	defer rows.Close()

	var id, fullURL string
	var deleted bool
	for rows.Next() {
		err = rows.Scan(&id, &fullURL, &deleted)

		if err != nil {
			return domain.URLForAddStorage{}, err
		}

	}
	err = rows.Err()
	if err != nil {
		return domain.URLForAddStorage{}, err
	}

	return domain.URLForAddStorage{
		FullURL: fullURL,
		Deleted: deleted,
	}, nil
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

func (f DBStorage) Ping(ctx context.Context) error {
	if err := f.db.Ping(ctx); err != nil {
		log.Print("ping error")
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
