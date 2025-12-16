package model

import (
	"net/http"
)

var (
	ErrPublisherAlreadyExists = NewInternalError(http.StatusConflict, "publisher already exists")
	ErrPublisherNotFound      = NewInternalError(http.StatusNotFound, "publisher not found")
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
