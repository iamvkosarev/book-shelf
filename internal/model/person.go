package model

import (
	"github.com/google/uuid"
)

type Person struct {
	ID         uuid.UUID
	FirstName  string
	LastName   string
	MiddleName string
}
