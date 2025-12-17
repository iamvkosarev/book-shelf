package handler

import (
	"encoding/json"
	"github.com/iamvkosarev/book-shelf/pkg/logs"
	"net/http"
)

const (
	VarID          = "id"
	VarExpend      = "expand"
	VarValuePerson = "person"
)

func decode[t interface{}](
	writer http.ResponseWriter,
	request *http.Request,
	requestData *t,
) bool {
	if err := json.NewDecoder(request.Body).Decode(&requestData); err != nil {
		sendError(writer, err)
		logs.Error("failed to decode the request", err)
		return false
	}
	return true
}
