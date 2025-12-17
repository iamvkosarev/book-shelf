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
	InvalidGenreID = "invalid genre id"
	MissingGenreID = "missing genre id"
)

type GenreUsecase interface {
	AddGenre(ctx context.Context, name string) (uuid.UUID, error)
	GetGenre(ctx context.Context, id uuid.UUID) (model.Genre, error)
	UpdateGenre(ctx context.Context, id uuid.UUID, name string) error
	RemoveGenre(ctx context.Context, id uuid.UUID) error
	ListGenres(ctx context.Context) ([]model.Genre, error)
}

type GenreHandler struct {
	genreUsecase GenreUsecase
	validate     *validator.Validate
}

func NewGenreHandler(usecase GenreUsecase) *GenreHandler {
	return &GenreHandler{
		genreUsecase: usecase,
		validate:     validator.New(),
	}
}

type AddGenreRequest struct {
	Name string `json:"name" validate:"required,lte=50,gt=0"`
}

type AddGenreResponse struct {
	ID uuid.UUID `json:"id"`
}

func (p GenreHandler) AddGenre(writer http.ResponseWriter, request *http.Request) {
	var requestData AddGenreRequest
	if ok := decode(writer, request, &requestData); !ok {
		return
	}
	if validationErr(writer, p.validate, requestData) {
		return
	}
	id, err := p.genreUsecase.AddGenre(
		request.Context(), requestData.Name,
	)
	if err != nil {
		sendError(writer, err)
		logs.Error("failed to add genre", err)
		return
	}
	sendCreatedJSON(writer, AddGenreResponse{ID: id})
}

func (p GenreHandler) GetGenre(writer http.ResponseWriter, request *http.Request) {
	idStr := mux.Vars(request)[VarID]
	if idStr == "" {
		sendBadRequest(writer, MissingGenreID)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		sendBadRequest(writer, InvalidGenreID)
		return
	}
	genre, err := p.genreUsecase.GetGenre(request.Context(), id)
	if err != nil {
		sendError(writer, err)
		logs.Error("failed to get genre", err, slog.String("genre_id", idStr))
		return
	}

	response := GenreResponse{
		ID:   genre.ID,
		Name: genre.Name,
	}
	sendCreatedJSON(writer, response)
}

func (p GenreHandler) RemoveGenre(writer http.ResponseWriter, request *http.Request) {
	idStr := mux.Vars(request)[VarID]
	if idStr == "" {
		sendBadRequest(writer, MissingGenreID)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		sendBadRequest(writer, InvalidGenreID)
		return
	}

	if err = p.genreUsecase.RemoveGenre(request.Context(), id); err != nil {
		sendError(writer, err)
		logs.Error("failed to remove genre", err, slog.String("genre_id", idStr))
		return
	}
	sendOK(writer)
}

type GenreResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type ListGenresResponse struct {
	Genres []GenreResponse `json:"genres"`
}

func (p GenreHandler) ListGenres(writer http.ResponseWriter, request *http.Request) {
	genres, err := p.genreUsecase.ListGenres(
		request.Context(),
	)
	if err != nil {
		sendError(writer, err)
		logs.Error("failed to list genres", err)
		return
	}
	response := ListGenresResponse{
		make([]GenreResponse, len(genres)),
	}
	for i, genre := range genres {
		response.Genres[i] = GenreResponse{
			ID:   genre.ID,
			Name: genre.Name,
		}
	}

	sendCreatedJSON(writer, response)
}

type UpdateGenreRequest struct {
	Name string `json:"name" validate:"lte=50,gte=0"`
}

func (p GenreHandler) UpdateGenre(writer http.ResponseWriter, request *http.Request) {
	idStr := mux.Vars(request)[VarID]
	if idStr == "" {
		sendBadRequest(writer, MissingGenreID)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		sendBadRequest(writer, InvalidGenreID)
		return
	}
	var requestData UpdateGenreRequest
	if ok := decode(writer, request, &requestData); !ok {
		return
	}
	if validationErr(writer, p.validate, requestData) {
		return
	}
	if err = p.genreUsecase.UpdateGenre(request.Context(), id, requestData.Name); err != nil {
		sendError(writer, err)
		logs.Error("failed to update genre", err)
		return
	}
	sendOK(writer)
}
