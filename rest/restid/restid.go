package restid

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/forkyid/go-utils/aes"
)

// ID type for database/json id handling
type ID struct {
	Raw       uint
	Encrypted string
	Valid     bool
}

// MarshalJSON marshal
func (id ID) MarshalJSON() ([]byte, error) {
	return []byte(`"` + id.Encrypted + `"`), nil
}

// UnmarshalJSON unmarshal
func (id *ID) UnmarshalJSON(j []byte) error {
	err := json.Unmarshal(j, &(id.Encrypted))
	if err != nil {
		return err
	}
	decrypted := aes.Decrypt(id.Encrypted)
	id.Raw = uint(decrypted)
	id.Valid = true
	if decrypted < 0 {
		id.Raw = 0
		id.Valid = false
	}
	return nil
}

// Value implements valuer
func (id *ID) Value() (driver.Value, error) {
	if !id.Valid {
		return nil, nil
	}
	return id.Raw, nil
}

// Scan implements scanner
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

// IDFromRaw constructor from raw id
func IDFromRaw(raw uint) (id ID) {
	id.Raw = raw
	id.Valid = true
	id.Encrypted = aes.Encrypt(int(raw))
	return id
}

// IDFromEncrypted constructor from encrypted id
func IDFromEncrypted(encrypted string) (id ID) {
	id.Encrypted = encrypted
	id.Valid = true
	decrypted := aes.Decrypt(encrypted)
	id.Raw = uint(decrypted)
	id.Valid = true
	if decrypted < 0 {
		id.Raw = 0
		id.Valid = false
		return id
	}
	return id
}

// ArrayToRaw return raw id array
func ArrayToRaw(ids *[]ID) *[]uint {
	raws := make([]uint, len(*ids))
	for i := range *ids {
		raws[i] = (*ids)[i].Raw
	}
	return &raws
}

// ArrayToEncrypted return encrypted id array
func ArrayToEncrypted(ids *[]ID) *[]string {
	encypteds := make([]string, len(*ids))
	for i := range *ids {
		encypteds[i] = (*ids)[i].Encrypted
	}
	return &encypteds
}
