package uuid

import "github.com/google/uuid"

// GetUUID generates uuid as string
func GetUUID() string {
	return uuid.New().String()
}
