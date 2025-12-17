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
	PersonsHandler    *handler.PersonHandler
}

func Setup(rt *mux.Router, cfg config.Router, deps Deps) (http.Handler, error) {
	rt.Use(middleware.HandlePanic)
	rt.Use(
		func(next http.Handler) http.Handler {
			return http.TimeoutHandler(next, cfg.APITimeout, "request timeout")
		},
	)
	rt.Use(middleware.LogProcessesEdges)

	rt.HandleFunc("/publishers", deps.PublishersHandler.AddPublisher).Methods(http.MethodPost)
	rt.HandleFunc("/publishers", deps.PublishersHandler.ListPublishers).Methods(http.MethodGet)
	rt.HandleFunc("/publishers/{id}", deps.PublishersHandler.GetPublisher).Methods(http.MethodGet)
	rt.HandleFunc("/publishers/{id}", deps.PublishersHandler.UpdatePublisher).Methods(http.MethodPut)
	rt.HandleFunc("/publishers/{id}", deps.PublishersHandler.RemovePublisher).Methods(http.MethodDelete)

	rt.HandleFunc("/persons", deps.PersonsHandler.AddPerson).Methods(http.MethodPost)
	rt.HandleFunc("/persons", deps.PersonsHandler.ListPersons).Methods(http.MethodGet)
	rt.HandleFunc("/persons/{id}", deps.PersonsHandler.GetPerson).Methods(http.MethodGet)
	rt.HandleFunc("/persons/{id}", deps.PersonsHandler.UpdatePerson).Methods(http.MethodPut)
	rt.HandleFunc("/persons/{id}", deps.PersonsHandler.RemovePerson).Methods(http.MethodDelete)

	rt.HandleFunc("/authors", deps.AuthorsHandler.AddAuthor).Methods(http.MethodPost)
	rt.HandleFunc("/authors", deps.AuthorsHandler.ListAuthors).Methods(http.MethodGet)
	rt.HandleFunc("/authors/{id}", deps.AuthorsHandler.GetAuthor).Methods(http.MethodGet)
	rt.HandleFunc("/authors/{id}", deps.AuthorsHandler.UpdateAuthor).Methods(http.MethodPut)
	rt.HandleFunc("/authors/{id}", deps.AuthorsHandler.RemoveAuthor).Methods(http.MethodDelete)

	return rt, nil
}
