package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	Created_at  time.Time `json:"created_at"`
	Updated_at  time.Time `json:"updated_at"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}
