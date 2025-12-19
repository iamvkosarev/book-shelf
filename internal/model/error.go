package model

import (
	"net/http"
)

var (
	ErrPublisherAlreadyExists = NewInternalError(http.StatusConflict, "publisher already exists")
	ErrPublisherNotFound      = NewInternalError(http.StatusNotFound, "publisher not found")

	ErrAuthorInvalidFields = NewInternalError(
		http.StatusBadRequest, "author must have first_name, "+
			"last_name or pseudonym",
	)
	ErrAuthorAlreadyExists = NewInternalError(http.StatusConflict, "author already exists")
	ErrAuthorNotFound      = NewInternalError(http.StatusNotFound, "author not found")

	ErrTagAlreadyExists = NewInternalError(http.StatusConflict, "tag already exists")
	ErrTagNotFound      = NewInternalError(http.StatusNotFound, "tag not found")

	ErrBookAlreadyExists = NewInternalError(http.StatusConflict, "book already exists")
	ErrBookNotFound      = NewInternalError(http.StatusNotFound, "book not found")
	ErrBookInvalidFields = NewInternalError(http.StatusBadRequest, "book invalid fields")
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
