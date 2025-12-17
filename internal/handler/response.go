package handler

import (
	"encoding/json"
	"errors"
	"github.com/iamvkosarev/book-shelf/internal/model"
	"net/http"
)

func sendError(writer http.ResponseWriter, err error) {
	var internalError *model.InternalError
	code := http.StatusInternalServerError
	message := "internal server error"

	if errors.As(err, &internalError) {
		code = internalError.Code()
		message = internalError.Error()
	}
	sendErrorJSON(writer, code, message)
}

func sendErrorJSON(writer http.ResponseWriter, code int, message string) {
	errorData := struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}{
		Code:    code,
		Message: message,
	}
	sendJSON(writer, errorData, errorData.Code)
}

func sendBadRequest(writer http.ResponseWriter, message string) {
	sendErrorJSON(writer, http.StatusBadRequest, message)
}

func sendCreatedJSON(writer http.ResponseWriter, response any) {
	sendJSON(writer, response, http.StatusCreated)
}

func sendOk(writer http.ResponseWriter) {
	writer.WriteHeader(http.StatusOK)
}

func sendOkJSON(writer http.ResponseWriter, response any) {
	sendJSON(writer, response, http.StatusOK)
}

func sendJSON(writer http.ResponseWriter, response any, status int) {
	body, _ := json.Marshal(response)
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	writer.Write(body)
}
