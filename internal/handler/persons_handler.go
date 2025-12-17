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
	InvalidPersonID = "invalid person id"
	MissingPersonID = "missing person id"
)

type PersonUsecase interface {
	AddPerson(ctx context.Context, firstName, lastName, middleName string) (uuid.UUID, error)
	GetPerson(ctx context.Context, id uuid.UUID) (model.Person, error)
	UpdatePerson(ctx context.Context, id uuid.UUID, firstName, lastName, middleName string) error
	RemovePerson(ctx context.Context, id uuid.UUID) error
	ListPersons(ctx context.Context) ([]model.Person, error)
}

type PersonHandler struct {
	personUsecase PersonUsecase
	validate      *validator.Validate
}

func NewPersonHandler(usecase PersonUsecase) *PersonHandler {
	return &PersonHandler{
		personUsecase: usecase,
		validate:      validator.New(),
	}
}

type AddPersonRequest struct {
	FirstName  string `json:"first_name" validate:"required,min=1,max=100"`
	LastName   string `json:"last_name" validate:"required,min=1,max=100"`
	MiddleName string `json:"middle_name" validate:"required,min=0,max=100"`
}

type AddPersonResponse struct {
	ID uuid.UUID `json:"id"`
}

func (p PersonHandler) AddPerson(writer http.ResponseWriter, request *http.Request) {
	var requestData AddPersonRequest
	if ok := decode(writer, request, &requestData); !ok {
		return
	}
	if validationErr(writer, p.validate, requestData) {
		return
	}
	id, err := p.personUsecase.AddPerson(
		request.Context(), requestData.FirstName, requestData.LastName, requestData.MiddleName,
	)
	if err != nil {
		sendError(writer, err)
		logs.Error("failed to add person", err)
		return
	}
	sendCreatedJSON(writer, AddPersonResponse{ID: id})
}

type PersonResponse struct {
	ID         uuid.UUID `json:"id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	MiddleName string    `json:"middle_name"`
}

func (p PersonHandler) GetPerson(writer http.ResponseWriter, request *http.Request) {
	idStr := mux.Vars(request)[VarID]
	if idStr == "" {
		sendBadRequest(writer, MissingPersonID)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		sendBadRequest(writer, InvalidPersonID)
		return
	}
	person, err := p.personUsecase.GetPerson(request.Context(), id)
	if err != nil {
		sendError(writer, err)
		logs.Error("failed to get person", err, slog.String("person_id", idStr))
		return
	}

	response := PersonResponse{
		ID:         person.ID,
		FirstName:  person.FirstName,
		LastName:   person.LastName,
		MiddleName: person.MiddleName,
	}
	sendCreatedJSON(writer, response)
}

func (p PersonHandler) RemovePerson(writer http.ResponseWriter, request *http.Request) {
	idStr := mux.Vars(request)[VarID]
	if idStr == "" {
		sendBadRequest(writer, MissingPersonID)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		sendBadRequest(writer, InvalidPersonID)
		return
	}

	if err = p.personUsecase.RemovePerson(request.Context(), id); err != nil {
		sendError(writer, err)
		logs.Error("failed to remove person", err, slog.String("person_id", idStr))
		return
	}
	sendOK(writer)
}

type ListPersonsResponse struct {
	Persons []PersonResponse `json:"persons"`
}

func (p PersonHandler) ListPersons(writer http.ResponseWriter, request *http.Request) {
	persons, err := p.personUsecase.ListPersons(
		request.Context(),
	)
	if err != nil {
		sendError(writer, err)
		logs.Error("failed to list persons", err)
		return
	}
	response := ListPersonsResponse{
		make([]PersonResponse, len(persons)),
	}
	for i, person := range persons {
		response.Persons[i] = PersonResponse{
			ID:         person.ID,
			FirstName:  person.FirstName,
			LastName:   person.LastName,
			MiddleName: person.MiddleName,
		}
	}

	sendCreatedJSON(writer, response)
}

type UpdatePersonRequest struct {
	FirstName  string `json:"first_name" validate:"min=1,max=100"`
	LastName   string `json:"last_name" validate:"min=1,max=100"`
	MiddleName string `json:"middle_name" validate:"min=0,max=100"`
}

func (p PersonHandler) UpdatePerson(writer http.ResponseWriter, request *http.Request) {
	idStr := mux.Vars(request)[VarID]
	if idStr == "" {
		sendBadRequest(writer, MissingPersonID)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		sendBadRequest(writer, InvalidPersonID)
		return
	}
	var requestData UpdatePersonRequest
	if ok := decode(writer, request, &requestData); !ok {
		return
	}
	if validationErr(writer, p.validate, requestData) {
		return
	}
	if err = p.personUsecase.UpdatePerson(
		request.Context(), id, requestData.FirstName, requestData.LastName, requestData.MiddleName,
	); err != nil {
		sendError(writer, err)
		logs.Error("failed to update person", err)
		return
	}
	sendOK(writer)
}
