package model

import (
	"github.com/google/uuid"
	"time"
)

type Book struct {
	ID          uuid.UUID
	PublisherID *uuid.UUID
	AuthorsIDs  []uuid.UUID
	TagsIDs     []uuid.UUID
	PublishedAt *time.Time
	Title       string
	Description *string
	Price       *float64
	Mark        *int16
	Publisher   *Publisher
	Authors     []Author
	Tags        []Tag
}
