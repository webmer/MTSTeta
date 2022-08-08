package uuid

import (
	"github.com/google/uuid"
)

func GenUUID() string {
	id := uuid.New()
	return id.String()
}
