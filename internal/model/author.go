package model

import (
	"github.com/google/uuid"
)

type Author struct {
	ID        uuid.UUID
	PersonID  uuid.UUID
	Pseudonym string
	Person    Person
}
