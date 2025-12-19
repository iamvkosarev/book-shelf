package model

import (
	"net/http"
)

var (
	ErrPublisherAlreadyExists = NewInternalError(http.StatusConflict, "publisher already exists")
	ErrPublisherNotFound      = NewInternalError(http.StatusNotFound, "publisher not found")

	ErrAuthorInvalidFields = NewInternalError(
		http.StatusBadRequest, "author must have first_name, last_name or pseudonym",
	)
	ErrAuthorAlreadyExists = NewInternalError(http.StatusConflict, "author already exists")
	ErrAuthorNotFound      = NewInternalError(http.StatusNotFound, "author not found")

	ErrTagAlreadyExists = NewInternalError(http.StatusConflict, "tag already exists")
	ErrTagNotFound      = NewInternalError(http.StatusNotFound, "tag not found")

	ErrBookAlreadyExists = NewInternalError(http.StatusConflict, "book already exists")
	ErrBookNotFound      = NewInternalError(http.StatusNotFound, "book not found")
	ErrBookInvalidFields = NewInternalError(http.StatusBadRequest, "book invalid fields")

	ErrPasswordTooShort = NewInternalError(http.StatusBadRequest, "password is too short")
	ErrUserNotExists    = NewInternalError(
		http.StatusBadRequest,
		"password or user name is not correct",
	)
	ErrPasswordNotCorrect = NewInternalError(
		http.StatusBadRequest, "password or user name is not correct",
	)
	ErrNoUpperCase = NewInternalError(
		http.StatusBadRequest,
		"password should contains at least one uppercase latter",
	)
	ErrNoLowerCase = NewInternalError(
		http.StatusBadRequest,
		"password should contains at least one lowercase latter",
	)
	ErrNoLetter = NewInternalError(
		http.StatusBadRequest,
		"password should contains at least one letter",
	)
	ErrNoNumber = NewInternalError(
		http.StatusBadRequest,
		"password should contains at least one number",
	)
	ErrUserRoleHasNoAccess = NewInternalError(
		http.StatusForbidden,
		"has no access",
	)
	ErrUserAlreadyExists = NewInternalError(http.StatusConflict, "user already exists")
	ErrUserNotFound      = NewInternalError(http.StatusConflict, "user not found")

	ErrTokenNotFound = NewInternalError(
		http.StatusBadRequest,
		"token not found",
	)
	ErrSignatureInvalid = NewInternalError(
		http.StatusUnauthorized,
		"token signature is invalid",
	)
	ErrTokenExpired = NewInternalError(
		http.StatusUnauthorized,
		"token is expired",
	)
	ErrTokenVerification = NewInternalError(
		http.StatusUnauthorized,
		"token signature is invalid",
	)
	ErrTokenDecryption = NewInternalError(
		http.StatusUnauthorized,
		"failed to decrypt token",
	)
	ErrParseClaims = NewInternalError(
		http.StatusInternalServerError,
		"failed to parse claims",
	)
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
