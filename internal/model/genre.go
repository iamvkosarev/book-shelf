package model

import (
	"github.com/google/uuid"
)

type Genre struct {
	ID   uuid.UUID
	Name string
}
