package handler

import (
	"context"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/iamvkosarev/book-shelf/internal/model"
	"github.com/iamvkosarev/book-shelf/internal/usecase/books"
	"github.com/iamvkosarev/book-shelf/pkg/logs"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

const (
	InvalidBookID = "invalid book id"
	MissingBookID = "missing book id"
)

type BookUsecase interface {
	AddBook(ctx context.Context, input books.CreateBookInput) (uuid.UUID, error)
	GetBook(
		ctx context.Context,
		id uuid.UUID,
		expendAuthorsData bool,
		expendTagsData bool,
		expendPublisherData bool,
	) (model.Book, error)
	UpdateBook(ctx context.Context, id uuid.UUID, patch books.UpdateBookPatch) error
	RemoveBook(ctx context.Context, id uuid.UUID) error
	ListBooks(
		ctx context.Context,
		parameters books.ListBookParameters,
		expandAuthors, expandTags, expandPublisher bool,
	) ([]model.Book, error)
}

type BookHandler struct {
	bookUsecase BookUsecase
	validate    *validator.Validate
}

func NewBookHandler(usecase BookUsecase) *BookHandler {
	return &BookHandler{
		bookUsecase: usecase,
		validate:    validator.New(),
	}
}

type AddBookRequest struct {
	PublisherID *uuid.UUID  `json:"publisher_id"`
	AuthorsIDs  []uuid.UUID `json:"authors_ids" validate:"dive,required"`
	TagsIDs     []uuid.UUID `json:"tags_ids" validate:"dive,required"`
	PublishedAt *time.Time  `json:"published_at"`
	Title       string      `json:"title" validate:"required,min=1,max=100"`
	Description *string     `json:"description" validate:"omitempty,max=1000"`
	Price       *float64    `json:"price"`
}

type AddBookResponse struct {
	ID uuid.UUID `json:"id"`
}

func (p BookHandler) AddBook(writer http.ResponseWriter, request *http.Request) {
	var requestData AddBookRequest
	if ok := decode(writer, request, &requestData); !ok {
		return
	}
	if validationErr(writer, p.validate, requestData) {
		return
	}
	input := books.CreateBookInput{
		AuthorsIDs:  requestData.AuthorsIDs,
		PublisherID: requestData.PublisherID,
		TagsIDs:     requestData.TagsIDs,
		PublishedAt: requestData.PublishedAt,
		Title:       requestData.Title,
		Description: requestData.Description,
		Price:       requestData.Price,
	}

	id, err := p.bookUsecase.AddBook(request.Context(), input)
	if err != nil {
		sendError(writer, err)
		logs.Error("failed to add book", err)
		return
	}
	sendCreatedJSON(writer, AddBookResponse{ID: id})
}

type BookResponse struct {
	ID          uuid.UUID          `json:"id"`
	PublisherID *uuid.UUID         `json:"publisher_id"`
	AuthorsIDs  []uuid.UUID        `json:"authors_ids"`
	TagsIDs     []uuid.UUID        `json:"tags_ids"`
	PublishedAt *time.Time         `json:"published_at"`
	Title       string             `json:"title"`
	Description *string            `json:"description"`
	Price       *float64           `json:"price"`
	Publisher   *PublisherResponse `json:"publisher"`
	Authors     []AuthorResponse   `json:"authors"`
	Tags        []TagResponse      `json:"tags"`
}

func (p BookHandler) GetBook(writer http.ResponseWriter, request *http.Request) {
	idStr := mux.Vars(request)[VarID]
	if idStr == "" {
		sendBadRequest(writer, MissingBookID)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		sendBadRequest(writer, InvalidBookID)
		return
	}

	expend := parseQueryToStringMap(request, VarExpend)
	expendAuthorsData := hasKeyInMap(expend, VarExpendValueAuthors)
	expendTagsData := hasKeyInMap(expend, VarExpendValueTags)
	expendPublisherData := hasKeyInMap(expend, VarExpendValuePublisher)

	book, err := p.bookUsecase.GetBook(request.Context(), id, expendAuthorsData, expendTagsData, expendPublisherData)
	if err != nil {
		sendError(writer, err)
		logs.Error("failed to get book", err, slog.String("book_id", idStr))
		return
	}

	response := getBookResponse(book, expendAuthorsData, expendTagsData, expendPublisherData)

	sendOkJSON(writer, response)
}

func (p BookHandler) RemoveBook(writer http.ResponseWriter, request *http.Request) {
	idStr := mux.Vars(request)[VarID]
	if idStr == "" {
		sendBadRequest(writer, MissingBookID)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		sendBadRequest(writer, InvalidBookID)
		return
	}

	if err = p.bookUsecase.RemoveBook(request.Context(), id); err != nil {
		sendError(writer, err)
		logs.Error("failed to remove book", err, slog.String("book_id", idStr))
		return
	}
	sendOk(writer)
}

type ListBooksResponse struct {
	Books []BookResponse `json:"books"`
}

func (p BookHandler) ListBooks(writer http.ResponseWriter, request *http.Request) {
	expend := parseQueryToStringMap(request, VarExpend)
	expendAuthorsData := hasKeyInMap(expend, VarExpendValueAuthors)
	expendTagsData := hasKeyInMap(expend, VarExpendValueTags)
	expendPublisherData := hasKeyInMap(expend, VarExpendValuePublisher)

	authors, err := parseQueryToUUIDs(request, VarAuthorID)
	if err != nil {
		sendError(writer, err)
		return
	}
	authorsIDs := make([]string, 0, len(authors))
	for id := range authors {
		authorsIDs = append(authorsIDs, authors[id].String())
	}
	slog.Info("search for authors", slog.String("authors_ids", strings.Join(authorsIDs, ",")))

	parameters := books.ListBookParameters{
		AuthorsIDs: authors,
	}

	books, err := p.bookUsecase.ListBooks(
		request.Context(), parameters,
		expendAuthorsData, expendTagsData, expendPublisherData,
	)
	if err != nil {
		sendError(writer, err)
		logs.Error("failed to list books", err)
		return
	}
	response := ListBooksResponse{
		make([]BookResponse, len(books)),
	}
	for i, book := range books {
		response.Books[i] = getBookResponse(book, expendAuthorsData, expendTagsData, expendPublisherData)
	}

	sendOkJSON(writer, response)
}

type UpdateBookRequest struct {
	PublisherID *uuid.UUID  `json:"publisher_id"`
	AuthorsIDs  []uuid.UUID `json:"authors_ids"`
	TagsIDs     []uuid.UUID `json:"tags_ids"`
	PublishedAt *time.Time  `json:"published_at"`
	Title       *string     `json:"title" validate:"omitempty,min=1,max=100"`
	Description *string     `json:"description" validate:"omitempty,max=1000"`
	Price       *float64    `json:"price"`
}

func (p BookHandler) UpdateBook(writer http.ResponseWriter, request *http.Request) {
	idStr := mux.Vars(request)[VarID]
	if idStr == "" {
		sendBadRequest(writer, MissingBookID)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		sendBadRequest(writer, InvalidBookID)
		return
	}
	var requestData UpdateBookRequest
	if ok := decode(writer, request, &requestData); !ok {
		return
	}
	if validationErr(writer, p.validate, requestData) {
		return
	}
	patch := books.UpdateBookPatch{
		PublisherID: requestData.PublisherID,
		AuthorsIDs:  requestData.AuthorsIDs,
		TagsIDs:     requestData.TagsIDs,
		PublishedAt: requestData.PublishedAt,
		Title:       requestData.Title,
		Description: requestData.Description,
		Price:       requestData.Price,
	}
	if err = p.bookUsecase.UpdateBook(request.Context(), id, patch); err != nil {
		sendError(writer, err)
		logs.Error("failed to update book", err)
		return
	}
	sendOk(writer)
}

func getBookResponse(
	book model.Book,
	expendAuthorsData bool,
	expendTagsData bool,
	expendPublisherData bool,
) BookResponse {
	response := BookResponse{
		ID:          book.ID,
		PublisherID: book.PublisherID,
		AuthorsIDs:  book.AuthorsIDs,
		TagsIDs:     book.TagsIDs,
		PublishedAt: book.PublishedAt,
		Title:       book.Title,
		Description: book.Description,
		Price:       book.Price,
	}

	if expendPublisherData && book.PublisherID != nil && *book.PublisherID != uuid.Nil {
		response.Publisher = &PublisherResponse{
			ID:   book.Publisher.ID,
			Name: book.Publisher.Name,
		}
	}

	if expendAuthorsData && book.Authors != nil && len(book.Authors) > 0 {
		authorsResponse := make([]AuthorResponse, len(book.Authors))
		for i, author := range book.Authors {
			authorResponse := AuthorResponse{
				ID:         author.ID,
				FirstName:  author.FirstName,
				LastName:   author.LastName,
				MiddleName: author.MiddleName,
				Pseudonym:  author.Pseudonym,
			}
			authorsResponse[i] = authorResponse
		}
		response.Authors = authorsResponse
	}

	if expendTagsData && book.Tags != nil && len(book.Tags) > 0 {
		tags := make([]TagResponse, len(book.Tags))
		for i, tag := range book.Tags {
			tags[i] = TagResponse{
				ID:   tag.ID,
				Name: tag.Name,
			}
		}
		response.Tags = tags
	}
	return response
}
