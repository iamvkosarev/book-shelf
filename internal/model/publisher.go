package model

import (
	"github.com/google/uuid"
)

type Publisher struct {
	ID   uuid.UUID
	Name string
}
