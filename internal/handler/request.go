package handler

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/iamvkosarev/book-shelf/internal/model"
	"github.com/iamvkosarev/book-shelf/pkg/logs"
	"net/http"
	"strings"
)

const (
	VarID                   = "id"
	VarExpend               = "expand"
	VarAuthorID             = "author_id"
	VarExpendValuePerson    = "person"
	VarExpendValueAuthors   = "authors"
	VarExpendValuePublisher = "publisher"
	VarExpendValueTags      = "tags"
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

func parseQueryToUUIDs(request *http.Request, key string) ([]uuid.UUID, error) {
	raw := request.URL.Query().Get(key)
	if raw == "" {
		return nil, nil
	}
	partsMap := parseQueryToStringMap(request, key)
	result := make([]uuid.UUID, 0, len(partsMap))
	for part := range partsMap {
		id, err := uuid.Parse(part)
		if err != nil {
			return nil, model.NewInternalError(http.StatusBadRequest, fmt.Sprintf("invalid %s", key))
		}
		result = append(result, id)
	}
	return result, nil
}

func parseQueryToStringMap(r *http.Request, key string) map[string]struct{} {
	raw := r.URL.Query().Get(key)
	result := make(map[string]struct{})
	if raw == "" {
		return result
	}
	for _, part := range strings.Split(raw, ",") {
		trimPart := strings.TrimSpace(part)
		if trimPart != "" {
			result[trimPart] = struct{}{}
		}
	}
	return result
}

func hasKeyInMap(mp map[string]struct{}, key string) bool {
	_, ok := mp[key]
	return ok
}
