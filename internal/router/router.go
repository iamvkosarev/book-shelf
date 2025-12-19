package router

import (
	"github.com/gorilla/mux"
	"github.com/iamvkosarev/book-shelf/config"
	"github.com/iamvkosarev/book-shelf/internal/handler"
	"github.com/iamvkosarev/book-shelf/internal/router/middleware"
	"net/http"
)

type Deps struct {
	PublishersHandler *handler.PublisherHandler
	AuthorsHandler    *handler.AuthorHandler
	TagsHandler       *handler.TagHandler
	BooksHandler      *handler.BookHandler
}

func Setup(rt *mux.Router, cfg config.Router, deps Deps) (http.Handler, error) {
	rt.Use(middleware.HandlePanic)
	rt.Use(
		func(next http.Handler) http.Handler {
			return http.TimeoutHandler(next, cfg.APITimeout, "request timeout")
		},
	)
	rt.Use(middleware.LogProcessesEdges)

	rt.HandleFunc("/healthz", HealthHandler).Methods(http.MethodGet)

	rt.HandleFunc("/publishers", deps.PublishersHandler.AddPublisher).Methods(http.MethodPost)
	rt.HandleFunc("/publishers", deps.PublishersHandler.ListPublishers).Methods(http.MethodGet)
	rt.HandleFunc("/publishers/{id}", deps.PublishersHandler.GetPublisher).Methods(http.MethodGet)
	rt.HandleFunc("/publishers/{id}", deps.PublishersHandler.UpdatePublisher).Methods(http.MethodPut)
	rt.HandleFunc("/publishers/{id}", deps.PublishersHandler.RemovePublisher).Methods(http.MethodDelete)

	rt.HandleFunc("/authors", deps.AuthorsHandler.AddAuthor).Methods(http.MethodPost)
	rt.HandleFunc("/authors", deps.AuthorsHandler.ListAuthors).Methods(http.MethodGet)
	rt.HandleFunc("/authors/{id}", deps.AuthorsHandler.GetAuthor).Methods(http.MethodGet)
	rt.HandleFunc("/authors/{id}", deps.AuthorsHandler.UpdateAuthor).Methods(http.MethodPut)
	rt.HandleFunc("/authors/{id}", deps.AuthorsHandler.RemoveAuthor).Methods(http.MethodDelete)

	rt.HandleFunc("/tags", deps.TagsHandler.AddTag).Methods(http.MethodPost)
	rt.HandleFunc("/tags", deps.TagsHandler.ListTags).Methods(http.MethodGet)
	rt.HandleFunc("/tags/{id}", deps.TagsHandler.GetTag).Methods(http.MethodGet)
	rt.HandleFunc("/tags/{id}", deps.TagsHandler.UpdateTag).Methods(http.MethodPut)
	rt.HandleFunc("/tags/{id}", deps.TagsHandler.RemoveTag).Methods(http.MethodDelete)

	rt.HandleFunc("/books", deps.BooksHandler.AddBook).Methods(http.MethodPost)
	rt.HandleFunc("/books", deps.BooksHandler.ListBooks).Methods(http.MethodGet)
	rt.HandleFunc("/books/{id}", deps.BooksHandler.GetBook).Methods(http.MethodGet)
	rt.HandleFunc("/books/{id}", deps.BooksHandler.UpdateBook).Methods(http.MethodPut)
	rt.HandleFunc("/books/{id}", deps.BooksHandler.RemoveBook).Methods(http.MethodDelete)

	return rt, nil
}

func HealthHandler(writer http.ResponseWriter, _ *http.Request) {
	writer.WriteHeader(http.StatusOK)
}
