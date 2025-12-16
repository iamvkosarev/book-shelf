package router

import (
	"github.com/gorilla/mux"
	"github.com/iamvkosarev/book-shelf/config"
	"github.com/iamvkosarev/book-shelf/internal/handler"
	"github.com/iamvkosarev/book-shelf/internal/router/middleware"
	"net/http"
)

type Deps struct {
	PublisherHandler *handler.PublisherHandler
}

func Setup(rt *mux.Router, deps Deps, cfg config.Router) (http.Handler, error) {
	rt.Use(middleware.HandlePanic)
	rt.Use(
		func(next http.Handler) http.Handler {
			return http.TimeoutHandler(next, cfg.APITimeout, "request timeout")
		},
	)
	rt.Use(middleware.LogProcessesEdges)

	rt.HandleFunc("/publishers", deps.PublisherHandler.AddPublisher).Methods(http.MethodPost)
	rt.HandleFunc("/publishers", deps.PublisherHandler.ListPublishers).Methods(http.MethodGet)
	rt.HandleFunc("/publishers/{id}", deps.PublisherHandler.GetPublisher).Methods(http.MethodGet)
	rt.HandleFunc("/publishers/{id}", deps.PublisherHandler.UpdatePublisher).Methods(http.MethodPut)
	rt.HandleFunc("/publishers/{id}", deps.PublisherHandler.RemovePublisher).Methods(http.MethodDelete)

	return rt, nil
}
