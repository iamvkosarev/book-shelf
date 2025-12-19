package handler

import (
	"context"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/iamvkosarev/book-shelf/internal/model"
	"github.com/iamvkosarev/book-shelf/pkg/logs"
	"log/slog"
	"net/http"
)

const (
	InvalidTagID = "invalid tag id"
	MissingTagID = "missing tag id"
)

type TagUsecase interface {
	AddTag(ctx context.Context, name string) (uuid.UUID, error)
	GetTag(ctx context.Context, id uuid.UUID) (model.Tag, error)
	UpdateTag(ctx context.Context, id uuid.UUID, name string) error
	RemoveTag(ctx context.Context, id uuid.UUID) error
	ListTags(ctx context.Context) ([]model.Tag, error)
}

type TagHandler struct {
	tagUsecase TagUsecase
	validate   *validator.Validate
}

func NewTagHandler(usecase TagUsecase) *TagHandler {
	return &TagHandler{
		tagUsecase: usecase,
		validate:   validator.New(),
	}
}

type AddTagRequest struct {
	Name string `json:"name" validate:"required,min=1,max=50"`
}

type AddTagResponse struct {
	ID uuid.UUID `json:"id"`
}

func (p TagHandler) AddTag(writer http.ResponseWriter, request *http.Request) {
	var requestData AddTagRequest
	if ok := decode(writer, request, &requestData); !ok {
		return
	}
	if validationErr(writer, p.validate, requestData) {
		return
	}
	id, err := p.tagUsecase.AddTag(
		request.Context(), requestData.Name,
	)
	if err != nil {
		SendError(writer, err)
		logs.Error("failed to add tag", err)
		return
	}
	sendCreatedJSON(writer, AddTagResponse{ID: id})
}

func (p TagHandler) GetTag(writer http.ResponseWriter, request *http.Request) {
	idStr := mux.Vars(request)[VarID]
	if idStr == "" {
		sendBadRequest(writer, MissingTagID)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		sendBadRequest(writer, InvalidTagID)
		return
	}
	tag, err := p.tagUsecase.GetTag(request.Context(), id)
	if err != nil {
		SendError(writer, err)
		logs.Error("failed to get tag", err, slog.String("tag_id", idStr))
		return
	}

	response := TagResponse{
		ID:   tag.ID,
		Name: tag.Name,
	}
	sendOkJSON(writer, response)
}

func (p TagHandler) RemoveTag(writer http.ResponseWriter, request *http.Request) {
	idStr := mux.Vars(request)[VarID]
	if idStr == "" {
		sendBadRequest(writer, MissingTagID)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		sendBadRequest(writer, InvalidTagID)
		return
	}

	if err = p.tagUsecase.RemoveTag(request.Context(), id); err != nil {
		SendError(writer, err)
		logs.Error("failed to remove tag", err, slog.String("tag_id", idStr))
		return
	}
	sendOk(writer)
}

type TagResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type ListTagsResponse struct {
	Tags []TagResponse `json:"tags"`
}

func (p TagHandler) ListTags(writer http.ResponseWriter, request *http.Request) {
	tags, err := p.tagUsecase.ListTags(
		request.Context(),
	)
	if err != nil {
		SendError(writer, err)
		logs.Error("failed to list tags", err)
		return
	}
	response := ListTagsResponse{
		make([]TagResponse, len(tags)),
	}
	for i, tag := range tags {
		response.Tags[i] = TagResponse{
			ID:   tag.ID,
			Name: tag.Name,
		}
	}

	sendOkJSON(writer, response)
}

type UpdateTagRequest struct {
	Name string `json:"name" validate:"required,min=1,max=50"`
}

func (p TagHandler) UpdateTag(writer http.ResponseWriter, request *http.Request) {
	idStr := mux.Vars(request)[VarID]
	if idStr == "" {
		sendBadRequest(writer, MissingTagID)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		sendBadRequest(writer, InvalidTagID)
		return
	}
	var requestData UpdateTagRequest
	if ok := decode(writer, request, &requestData); !ok {
		return
	}
	if validationErr(writer, p.validate, requestData) {
		return
	}
	if err = p.tagUsecase.UpdateTag(request.Context(), id, requestData.Name); err != nil {
		SendError(writer, err)
		logs.Error("failed to update tag", err)
		return
	}
	sendOk(writer)
}
