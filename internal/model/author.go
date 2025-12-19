package model

import (
	"github.com/google/uuid"
)

type Author struct {
	ID         uuid.UUID
	PersonID   uuid.UUID
	FirstName  *string
	LastName   *string
	MiddleName *string
	Pseudonym  *string
}
