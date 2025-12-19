package handler

import (
	"context"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/iamvkosarev/book-shelf/internal/model"
	"github.com/iamvkosarev/book-shelf/internal/usecase"
	"github.com/iamvkosarev/book-shelf/pkg/logs"
	"log/slog"
	"net/http"
)

const (
	InvalidAuthorID = "invalid author id"
	MissingAuthorID = "missing author id"
)

type AuthorUsecase interface {
	AddAuthor(ctx context.Context, input usecase.AddAuthorInput) (uuid.UUID, error)
	GetAuthor(ctx context.Context, id uuid.UUID) (model.Author, error)
	UpdateAuthor(ctx context.Context, id uuid.UUID, input usecase.UpdateAuthorInput) error
	RemoveAuthor(ctx context.Context, id uuid.UUID) error
	ListAuthors(ctx context.Context) ([]model.Author, error)
}

type AuthorHandler struct {
	authorUsecase AuthorUsecase
	validate      *validator.Validate
}

func NewAuthorHandler(usecase AuthorUsecase) *AuthorHandler {
	return &AuthorHandler{
		authorUsecase: usecase,
		validate:      validator.New(),
	}
}

type AddAuthorRequest struct {
	FirstName  *string `json:"first_name" validate:"omitempty,max=100"`
	LastName   *string `json:"last_name" validate:"omitempty,max=100"`
	MiddleName *string `json:"middle_name" validate:"omitempty,max=100"`
	Pseudonym  *string `json:"pseudonym" validate:"omitempty,max=100"`
}

type AddAuthorResponse struct {
	ID uuid.UUID `json:"id"`
}

func (p AuthorHandler) AddAuthor(writer http.ResponseWriter, request *http.Request) {
	var requestData AddAuthorRequest
	if ok := decode(writer, request, &requestData); !ok {
		return
	}
	if validationErr(writer, p.validate, requestData) {
		return
	}
	input := usecase.AddAuthorInput{
		FirstName:  requestData.FirstName,
		LastName:   requestData.LastName,
		MiddleName: requestData.MiddleName,
		Pseudonym:  requestData.Pseudonym,
	}
	id, err := p.authorUsecase.AddAuthor(request.Context(), input)
	if err != nil {
		SendError(writer, err)
		logs.Error("failed to add author", err)
		return
	}
	sendCreatedJSON(writer, AddAuthorResponse{ID: id})
}

type AuthorResponse struct {
	ID         uuid.UUID `json:"id"`
	FirstName  *string   `json:"first_name"`
	LastName   *string   `json:"last_name"`
	MiddleName *string   `json:"middle_name"`
	Pseudonym  *string   `json:"pseudonym"`
}

func (p AuthorHandler) GetAuthor(writer http.ResponseWriter, request *http.Request) {
	idStr := mux.Vars(request)[VarID]
	if idStr == "" {
		sendBadRequest(writer, MissingAuthorID)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		sendBadRequest(writer, InvalidAuthorID)
		return
	}

	author, err := p.authorUsecase.GetAuthor(request.Context(), id)
	if err != nil {
		SendError(writer, err)
		logs.Error("failed to get author", err, slog.String("author_id", idStr))
		return
	}

	response := AuthorResponse{
		ID:         author.ID,
		FirstName:  author.FirstName,
		LastName:   author.LastName,
		MiddleName: author.MiddleName,
		Pseudonym:  author.Pseudonym,
	}

	sendOkJSON(writer, response)
}

func (p AuthorHandler) RemoveAuthor(writer http.ResponseWriter, request *http.Request) {
	idStr := mux.Vars(request)[VarID]
	if idStr == "" {
		sendBadRequest(writer, MissingAuthorID)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		sendBadRequest(writer, InvalidAuthorID)
		return
	}

	if err = p.authorUsecase.RemoveAuthor(request.Context(), id); err != nil {
		SendError(writer, err)
		logs.Error("failed to remove author", err, slog.String("author_id", idStr))
		return
	}
	sendOk(writer)
}

type ListAuthorsResponse struct {
	Authors []AuthorResponse `json:"authors"`
}

func (p AuthorHandler) ListAuthors(writer http.ResponseWriter, request *http.Request) {
	authors, err := p.authorUsecase.ListAuthors(request.Context())
	if err != nil {
		SendError(writer, err)
		logs.Error("failed to list authors", err)
		return
	}
	response := ListAuthorsResponse{
		make([]AuthorResponse, len(authors)),
	}
	for i, author := range authors {
		response.Authors[i] = AuthorResponse{
			ID:         author.ID,
			FirstName:  author.FirstName,
			LastName:   author.LastName,
			MiddleName: author.MiddleName,
			Pseudonym:  author.Pseudonym,
		}
	}

	sendOkJSON(writer, response)
}

type UpdateAuthorRequest struct {
	FirstName  *string `json:"first_name" validate:"omitempty,max=100"`
	LastName   *string `json:"last_name" validate:"omitempty,max=100"`
	MiddleName *string `json:"middle_name" validate:"omitempty,max=100"`
	Pseudonym  *string `json:"pseudonym" validate:"omitempty,max=100"`
}

func (p AuthorHandler) UpdateAuthor(writer http.ResponseWriter, request *http.Request) {
	idStr := mux.Vars(request)[VarID]
	if idStr == "" {
		sendBadRequest(writer, MissingAuthorID)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		sendBadRequest(writer, InvalidAuthorID)
		return
	}
	var requestData UpdateAuthorRequest
	if ok := decode(writer, request, &requestData); !ok {
		return
	}
	if validationErr(writer, p.validate, requestData) {
		return
	}

	input := usecase.UpdateAuthorInput{
		FirstName:  requestData.FirstName,
		LastName:   requestData.LastName,
		MiddleName: requestData.MiddleName,
		Pseudonym:  requestData.Pseudonym,
	}

	if err = p.authorUsecase.UpdateAuthor(request.Context(), id, input); err != nil {
		SendError(writer, err)
		logs.Error("failed to update author", err, slog.String("author_id", idStr))
		return
	}
	sendOk(writer)
}
