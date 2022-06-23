package cookies

import (
	"crypto/aes"
	"crypto/rand"
	"fmt"
	"log"
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

func VerificationCookie(dst string) {

	k := ListKey[dst]
	id := make([]byte, aes.BlockSize)
	aesblock, err := aes.NewCipher([]byte(k))
	if err != nil {
		log.Print(err)
		return
	}
	aesblock.Decrypt(id, []byte(dst))
	log.Println()
}

func CookiesFunc() {
	src := []byte("Ключ от сердца") // данные, которые хотим зашифровать
	fmt.Printf("original: %s\n", src)

	// константа aes.BlockSize определяет размер блока и равна 16 байтам
	key, err := generateRandom(aes.BlockSize) // ключ шифрования
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	// получаем cipher.Block
	aesblock, err := aes.NewCipher(key)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	dst := make([]byte, aes.BlockSize) // зашифровываем
	aesblock.Encrypt(dst, src)
	fmt.Printf("encrypted: %x\n", dst)

	src2 := make([]byte, aes.BlockSize) // расшифровываем
	aesblock.Decrypt(src2, dst)
	fmt.Printf("decrypted: %s\n", src2)
}
