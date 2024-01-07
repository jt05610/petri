package petri

import (
	"github.com/google/uuid"
)

func ID() string {
	return uuid.New().String()
}
