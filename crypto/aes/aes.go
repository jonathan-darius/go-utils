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
	if err != nil {
		return -1
	}
	return d[0]
}

//DecryptBulk Function
func DecryptBulk(data []string) (ret []int, err error) {
	for _, d := range data {
		decrypted := Decrypt(d)
		if decrypted <= 0 {
			return nil, fmt.Errorf("Decrypt failed")
		}
		ret = append(ret, decrypted)
	}
	return ret, nil
}
