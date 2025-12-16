package handler

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/iamvkosarev/book-shelf/pkg/logs"
	"net/http"
	"strings"
)

const InternalErrorResponseText string = "internal server error"

func validationErr(w http.ResponseWriter, validate *validator.Validate, req interface{}) bool {
	if err := validate.Struct(req); err != nil {
		var validatorErr validator.ValidationErrors
		if errors.As(err, &validatorErr) {
			errs := make([]string, len(validatorErr))
			for i, fieldError := range validatorErr {
				if fieldError.Param() != "" {
					errs[i] = fmt.Sprintf(
						"failed to validate field '%s', because of tag '%s:%s'",
						strings.ToLower(fieldError.Field()),
						fieldError.Tag(), fieldError.Param(),
					)
				} else {
					errs[i] = fmt.Sprintf(
						"failed to validate field '%s', because of tag '%s'", strings.ToLower(fieldError.Field()),
						fieldError.Tag(),
					)
				}
				break
			}
			sendBadRequest(w, strings.Join(errs, "\n"))
		} else {
			sendBadRequest(w, err.Error())
		}
		logs.Error("failed to validate the request", err)
		return true
	}
	return false
}
