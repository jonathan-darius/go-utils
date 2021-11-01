package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/speps/go-hashids"
)

var hd *hashids.HashIDData
var salt string
var minLength int

func initialize() {
	if hd != nil {
		return
	}

	hd = hashids.NewData()
	salt = os.Getenv("AES_KEY")
	minLengthStr := os.Getenv("AES_MIN_LENGTH")

	if salt == "" || minLengthStr == "" {
		log.Println("aes: env not found: AES_KEY or AES_MIN_LENGTH")
	}

	minLength, _ = strconv.Atoi(minLengthStr)
}

// Encrypt Function
func Encrypt(id int) string {
	initialize()
	hd.Salt = salt
	hd.MinLength = minLength
	h, _ := hashids.NewWithData(hd)
	encoded, _ := h.Encode([]int{id})
	return encoded
}

// Decrypt Function
func Decrypt(data string) int {
	initialize()
	hd.Salt = salt
	hd.MinLength = minLength
	h, _ := hashids.NewWithData(hd)
	d, err := h.DecodeWithError(data)
	if err != nil || len(d) < 1 {
		return -1
	}
	return d[0]
}

// DecryptBulk Function
func DecryptBulk(data []string) (ret []int, err error) {
	ret = make([]int, len(data))
	for i := range data {
		decrypted := Decrypt(data[i])
		if decrypted <= 0 {
			return nil, fmt.Errorf("Decrypt failed")
		}
		ret[i] = decrypted
	}
	return ret, nil
}

// EncryptBulk Function
func EncryptBulk(data []int) (ret []string) {
	ret = make([]string, len(data))
	for i := range data {
		ret[i] = Encrypt(data[i])
	}
	return ret
}

// EncryptString ...
func EncryptString(data []byte) ([]byte, error) {
	block, _ := aes.NewCipher([]byte(os.Getenv("AES_STRING_KEY")))
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// DecryptString ...
func DecryptString(data []byte) ([]byte, error) {
	key := []byte(os.Getenv("AES_STRING_KEY"))
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("invalid")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

func initializeCMS() {
	if hd != nil {
		return
	}

	hd = hashids.NewData()
	salt = os.Getenv("AES_KEY_CMS")
	minLengthStr := os.Getenv("AES_MIN_LENGTH")

	if salt == "" || minLengthStr == "" {
		log.Println("aes: env not found: AES_KEY_CMS or AES_MIN_LENGTH")
	}

	minLength, _ = strconv.Atoi(minLengthStr)
}

// EncryptCMS Function
func EncryptCMS(id int) string {
	initializeCMS()
	hd.Salt = salt
	hd.MinLength = minLength
	h, _ := hashids.NewWithData(hd)
	encoded, _ := h.Encode([]int{id})
	return encoded
}

// DecryptCMS Function
func DecryptCMS(data string) int {
	initializeCMS()
	hd.Salt = salt
	hd.MinLength = minLength
	h, _ := hashids.NewWithData(hd)
	d, err := h.DecodeWithError(data)
	if err != nil || len(d) < 1 {
		return -1
	}
	return d[0]
}