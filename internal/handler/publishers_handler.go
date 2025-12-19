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
	InvalidPublisherID = "invalid publisher id"
	MissingPublisherID = "missing publisher id"
)

type PublisherUsecase interface {
	AddPublisher(ctx context.Context, name string) (uuid.UUID, error)
	GetPublisher(ctx context.Context, id uuid.UUID) (model.Publisher, error)
	UpdatePublisher(ctx context.Context, id uuid.UUID, name string) error
	RemovePublisher(ctx context.Context, id uuid.UUID) error
	ListPublishers(ctx context.Context) ([]model.Publisher, error)
}

type PublisherHandler struct {
	publisherUsecase PublisherUsecase
	validate         *validator.Validate
}

func NewPublisherHandler(usecase PublisherUsecase) *PublisherHandler {
	return &PublisherHandler{
		publisherUsecase: usecase,
		validate:         validator.New(),
	}
}

type AddPublisherRequest struct {
	Name string `json:"name" validate:"required,min=1,max=50"`
}

type AddPublisherResponse struct {
	ID uuid.UUID `json:"id"`
}

func (p PublisherHandler) AddPublisher(writer http.ResponseWriter, request *http.Request) {
	var requestData AddPublisherRequest
	if ok := decode(writer, request, &requestData); !ok {
		return
	}
	if validationErr(writer, p.validate, requestData) {
		return
	}
	id, err := p.publisherUsecase.AddPublisher(
		request.Context(), requestData.Name,
	)
	if err != nil {
		SendError(writer, err)
		logs.Error("failed to add publisher", err)
		return
	}
	sendCreatedJSON(writer, AddPublisherResponse{ID: id})
}

func (p PublisherHandler) GetPublisher(writer http.ResponseWriter, request *http.Request) {
	idStr := mux.Vars(request)[VarID]
	if idStr == "" {
		sendBadRequest(writer, MissingPublisherID)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		sendBadRequest(writer, InvalidPublisherID)
		return
	}
	publisher, err := p.publisherUsecase.GetPublisher(request.Context(), id)
	if err != nil {
		SendError(writer, err)
		logs.Error("failed to get publisher", err, slog.String("publisher_id", idStr))
		return
	}

	response := PublisherResponse{
		ID:   publisher.ID,
		Name: publisher.Name,
	}
	sendOkJSON(writer, response)
}

func (p PublisherHandler) RemovePublisher(writer http.ResponseWriter, request *http.Request) {
	idStr := mux.Vars(request)[VarID]
	if idStr == "" {
		sendBadRequest(writer, MissingPublisherID)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		sendBadRequest(writer, InvalidPublisherID)
		return
	}

	if err = p.publisherUsecase.RemovePublisher(request.Context(), id); err != nil {
		SendError(writer, err)
		logs.Error("failed to remove publisher", err, slog.String("publisher_id", idStr))
		return
	}
	sendOk(writer)
}

type PublisherResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type ListPublishersResponse struct {
	Publishers []PublisherResponse `json:"publishers"`
}

func (p PublisherHandler) ListPublishers(writer http.ResponseWriter, request *http.Request) {
	publishers, err := p.publisherUsecase.ListPublishers(
		request.Context(),
	)
	if err != nil {
		SendError(writer, err)
		logs.Error("failed to list publishers", err)
		return
	}
	response := ListPublishersResponse{
		make([]PublisherResponse, len(publishers)),
	}
	for i, publisher := range publishers {
		response.Publishers[i] = PublisherResponse{
			ID:   publisher.ID,
			Name: publisher.Name,
		}
	}

	sendOkJSON(writer, response)
}

type UpdatePublisherRequest struct {
	Name string `json:"name" validate:"required,min=1,max=50"`
}

func (p PublisherHandler) UpdatePublisher(writer http.ResponseWriter, request *http.Request) {
	idStr := mux.Vars(request)[VarID]
	if idStr == "" {
		sendBadRequest(writer, MissingPublisherID)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		sendBadRequest(writer, InvalidPublisherID)
		return
	}
	var requestData UpdatePublisherRequest
	if ok := decode(writer, request, &requestData); !ok {
		return
	}
	if validationErr(writer, p.validate, requestData) {
		return
	}
	if err = p.publisherUsecase.UpdatePublisher(request.Context(), id, requestData.Name); err != nil {
		SendError(writer, err)
		logs.Error("failed to update publisher", err)
		return
	}
	sendOk(writer)
}
