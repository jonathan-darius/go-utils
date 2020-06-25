package id

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/forkyid/go-utils/crypto/aes"
)

type ID struct {
	Raw       uint
	Encrypted string
	Valid     bool
}

func (id ID) MarshalJSON() ([]byte, error) {
	return []byte(`"` + id.Encrypted + `"`), nil
}

func (id *ID) UnmarshalJSON(j []byte) error {
	err := json.Unmarshal(j, &(id.Encrypted))
	if err != nil {
		return err
	}
	decrypted := aes.Decrypt(id.Encrypted)
	id.Raw = uint(decrypted)
	if decrypted < 0 {
		id.Raw = 0
		id.Valid = false
	}
	return nil
}

func (id *ID) Value() (driver.Value, error) {
	if !id.Valid {
		return nil, nil
	}
	return id.Raw, nil
}
func (id *ID) Scan(value interface{}) error {
	if value == nil {
		id.Valid = false
		return nil
	}

	// TODO: find a better way
	u, err := strconv.ParseUint(fmt.Sprint(value), 10, 64)
	if err != nil {
		return err
	}

	id.Raw = uint(u)
	id.Encrypted = aes.Encrypt(int(id.Raw))

	return nil
}

func IDFromRaw(raw uint) (id ID) {
	id.Raw = raw
	id.Valid = true
	id.Encrypted = aes.Encrypt(int(raw))
	return id
}

func IDFromEncrypted(encrypted string) (id ID) {
	id.Encrypted = encrypted
	id.Valid = true
	decrypted := aes.Decrypt(encrypted)
	id.Raw = uint(decrypted)
	if decrypted < 0 {
		id.Raw = 0
		id.Valid = false
		return id
	}
	return id
}
