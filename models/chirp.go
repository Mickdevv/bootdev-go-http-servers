package models

import (
	"time"

	"github.com/google/uuid"
)

type Chirp struct {
	Created_at time.Time `json:"created_at"`
	Updated_at time.Time `json:"updated_at"`
	User_id    uuid.UUID `json:"user_id"`
	Body       string    `json:"body"`
	Id         uuid.UUID `json:"id"`
}
