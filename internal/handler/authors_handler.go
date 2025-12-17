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
	InvalidAuthorID = "invalid author id"
	MissingAuthorID = "missing author id"
)

type AuthorUsecase interface {
	AddAuthor(
		ctx context.Context, personID uuid.UUID, pseudonym string,
	) (uuid.UUID, error)
	GetAuthor(ctx context.Context, id uuid.UUID, expendPersonData bool) (model.Author, error)
	UpdateAuthor(ctx context.Context, id, personID uuid.UUID, pseudonym string) error
	RemoveAuthor(ctx context.Context, id uuid.UUID) error
	ListAuthors(ctx context.Context, expendPersonData bool) ([]model.Author, error)
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
	PersonID  uuid.UUID `json:"person_id"`
	Pseudonym string    `json:"pseudonym" validate:"lte=100,gte=0"`
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
	id, err := p.authorUsecase.AddAuthor(request.Context(), requestData.PersonID, requestData.Pseudonym)
	if err != nil {
		sendError(writer, err)
		logs.Error("failed to add author", err)
		return
	}
	sendCreatedJSON(writer, AddAuthorResponse{ID: id})
}

type AuthorResponse struct {
	ID             uuid.UUID       `json:"id"`
	PersonID       uuid.UUID       `json:"person_id"`
	Pseudonym      string          `json:"pseudonym"`
	PersonResponse *PersonResponse `json:"person"`
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

	expendStr := request.URL.Query().Get(VarExpend)
	expendPersonData := expendStr == VarValuePerson

	author, err := p.authorUsecase.GetAuthor(request.Context(), id, expendPersonData)
	if err != nil {
		sendError(writer, err)
		logs.Error("failed to get author", err, slog.String("author_id", idStr))
		return
	}

	response := AuthorResponse{
		ID:        author.ID,
		PersonID:  author.PersonID,
		Pseudonym: author.Pseudonym,
	}

	if expendPersonData && author.PersonID != uuid.Nil {
		response.PersonResponse = &PersonResponse{
			ID:         author.Person.ID,
			FirstName:  author.Person.FirstName,
			LastName:   author.Person.LastName,
			MiddleName: author.Person.MiddleName,
		}
	}

	sendCreatedJSON(writer, response)
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
		sendError(writer, err)
		logs.Error("failed to remove author", err, slog.String("author_id", idStr))
		return
	}
	sendOK(writer)
}

type ListAuthorsResponse struct {
	Authors []AuthorResponse `json:"authors"`
}

func (p AuthorHandler) ListAuthors(writer http.ResponseWriter, request *http.Request) {
	expendStr := request.URL.Query().Get(VarExpend)
	expendPersonData := expendStr == VarValuePerson

	authors, err := p.authorUsecase.ListAuthors(request.Context(), expendPersonData)
	if err != nil {
		sendError(writer, err)
		logs.Error("failed to list authors", err)
		return
	}
	response := ListAuthorsResponse{
		make([]AuthorResponse, len(authors)),
	}
	for i, author := range authors {
		response.Authors[i] = AuthorResponse{
			ID:        author.ID,
			PersonID:  author.PersonID,
			Pseudonym: author.Pseudonym,
		}

		if expendPersonData && author.PersonID != uuid.Nil {
			response.Authors[i].PersonResponse = &PersonResponse{
				ID:         author.Person.ID,
				FirstName:  author.Person.FirstName,
				LastName:   author.Person.LastName,
				MiddleName: author.Person.MiddleName,
			}
		}
	}

	sendCreatedJSON(writer, response)
}

type UpdateAuthorRequest struct {
	PersonID  uuid.UUID `json:"person_id"`
	Pseudonym string    `json:"pseudonym"`
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
	if err = p.authorUsecase.UpdateAuthor(
		request.Context(), id, requestData.PersonID, requestData.Pseudonym,
	); err != nil {
		sendError(writer, err)
		logs.Error("failed to update author", err)
		return
	}
	sendOK(writer)
}
