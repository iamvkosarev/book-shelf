package model

import (
	"net/http"
)

var (
	ErrPublisherAlreadyExists = NewInternalError(http.StatusConflict, "publisher already exists")
	ErrPublisherNotFound      = NewInternalError(http.StatusNotFound, "publisher not found")

	ErrPersonAlreadyExists = NewInternalError(http.StatusConflict, "person already exists")
	ErrPersonNotFound      = NewInternalError(http.StatusNotFound, "person not found")

	ErrAuthorInvalidFields = NewInternalError(http.StatusBadRequest, "author must have person id or pseudonym")
	ErrAuthorAlreadyExists = NewInternalError(http.StatusConflict, "author already exists")
	ErrAuthorNotFound      = NewInternalError(http.StatusNotFound, "author not found")

	ErrGenreAlreadyExists = NewInternalError(http.StatusConflict, "genre already exists")
	ErrGenreNotFound      = NewInternalError(http.StatusNotFound, "genre not found")
)

type InternalError struct {
	code    int
	message string
}

func (r *InternalError) Error() string {
	return r.message
}

func (r *InternalError) Code() int {
	return r.code
}

func NewInternalError(code int, message string) *InternalError {
	return &InternalError{code, message}
}
