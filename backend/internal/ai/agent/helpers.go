package agent

import (
	"time"

	"github.com/google/uuid"
)

func generateID() string {
	return uuid.New().String()
}

func now() int64 {
	return time.Now().UnixMilli()
}
