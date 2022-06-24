package cookies

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"github.com/botaevg/yandexgo/internal/repositories"
	"log"
	"net/http"
	"strconv"
)

var ListKey map[string]string

func generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

var ID int = 0

func CreateCookie(s repositories.Storage) (string, string) {
	idStr := &ID
	*idStr++
	log.Print("новый ид пользователя: ")
	log.Print(*idStr)
	id := []byte(strconv.Itoa(*idStr))

	key, err := generateRandom(aes.BlockSize)
	if err != nil {
		log.Print(err)
		return "", ""
	}
	aesblock, err := aes.NewCipher(key)
	if err != nil {
		log.Print(err)
		return "", ""
	}
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		log.Print(err)
		return "", ""
	}
	nonce, err := generateRandom(aesgcm.NonceSize())
	if err != nil {
		return "", ""
	}

	dst := aesgcm.Seal(nil, nonce, id, nil)

	s.AddCookie(hex.EncodeToString(dst), key, nonce)
	log.Print("зашифрованный ид в виде строки: " + hex.EncodeToString(dst))

	return hex.EncodeToString(dst), string(id)
}

func DecryptID(s repositories.Storage, dst string) (string, error) {

	slID, err := s.GetID(dst)
	if err != nil {
		log.Print(err)
		return "", err
	}
	aesblock, err := aes.NewCipher((slID[0]))
	if err != nil {
		log.Print(err)
		return "", err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		log.Print(err)
		return "", err
	}

	x, err := hex.DecodeString(dst)
	if err != nil {
		log.Print(err)
		return "", err
	}
	id, err := aesgcm.Open(nil, (slID[1]), x, nil)
	if err != nil {

		log.Print(err)
		return "", err
	}
	return hex.EncodeToString(id), nil //hex.EncodeToString(id)
}

func VerificationCookie(h repositories.Storage, r *http.Request, w *http.ResponseWriter) string {

	x, err := r.Cookie("id_encrypt")
	if err != nil {
		log.Print("нет такого кука")
		valueEncrypt, idUser := CreateCookie(h)
		http.SetCookie(*w, &http.Cookie{
			Name:  "id_encrypt",
			Value: valueEncrypt,
		})
		return idUser
	} else {
		idUser, err := DecryptID(h, x.Value) //h.storage.GetId(x.Value)
		if err != nil {
			log.Print(err)
			log.Print("такой кук не найден, создадим новый")
			valueEncrypt, idUser := CreateCookie(h)
			log.Print("новый кук: " + valueEncrypt)
			http.SetCookie(*w, &http.Cookie{
				Name:  "id_encrypt",
				Value: valueEncrypt,
			})
			return idUser
		} else {
			log.Print("изначальный ИД: " + idUser)
			return idUser
		}
	}

}
