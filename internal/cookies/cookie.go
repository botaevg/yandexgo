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
	log.Print(*idStr)
	id := []byte(strconv.Itoa(*idStr))
	log.Print(id)

	key, err := generateRandom(aes.BlockSize)
	if err != nil {
		log.Print(err)
		log.Print("1")
		return "", ""
	}
	log.Print("2")
	aesblock, err := aes.NewCipher(key)

	aesgcm, err := cipher.NewGCM(aesblock)

	nonce, err := generateRandom(aesgcm.NonceSize())
	if err != nil {
		return "", ""
	}

	dst := aesgcm.Seal(nil, nonce, id, nil)
	//dst := make([]byte, aes.BlockSize)
	log.Print("3")
	log.Print(key)
	log.Print(nonce)

	//aesblock.Encrypt(dst, id)
	log.Print("4")

	//log.Print(dst)

	s.AddCookie(hex.EncodeToString(dst), key, nonce)
	log.Print("зашифрованный ид: ")
	log.Print(dst)
	log.Print("зашифрованный ид в виде строки: " + hex.EncodeToString(dst))

	return hex.EncodeToString(dst), string(id)
}

func DecryptID(s repositories.Storage, dst string) (string, error) {

	slId, err := s.GetId(dst)
	if err != nil {
		log.Print(err)
		return "", err
	}
	aesblock, err := aes.NewCipher((slId[0]))
	if err != nil {
		log.Print(err)
		return "", err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		log.Print(err)
		return "", err
	}
	log.Print("11")
	log.Print((slId[0]))
	log.Print((slId[1]))

	x, err := hex.DecodeString(dst)
	log.Print("5")
	log.Print(x)
	if err != nil {
		log.Print(err)
		log.Print("6")
		return "", err
	}
	id, err := aesgcm.Open(nil, (slId[1]), x, nil)
	log.Print("12")
	if err != nil {

		log.Print(err)
		log.Print(id)
		log.Print("13")
		return "", err
	}
	log.Print(string(id))
	return string(id), nil //hex.EncodeToString(id)
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
		log.Print(x)
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
