package aes

import (
	"fmt"
	"os"

	"github.com/speps/go-hashids"
)

var passphrase = os.Getenv("AES_KEY")
var hd = hashids.NewData()
var h, _ = hashids.NewWithData(hd)

// Encrypt Function
func Encrypt(id int) string {
	hd.Salt = passphrase
	hd.MinLength = 30
	hd = hashids.NewData()
	encoded, _ := h.Encode([]int{id})
	return encoded
}

// Decrypt Function
func Decrypt(data string) int {
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
