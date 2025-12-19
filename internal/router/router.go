package router

import (
	"context"
	"github.com/google/uuid"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/iamvkosarev/book-shelf/config"
	"github.com/iamvkosarev/book-shelf/internal/handler"
	"github.com/iamvkosarev/book-shelf/internal/model"
	"github.com/iamvkosarev/book-shelf/internal/router/middleware"
)

type Deps struct {
	PublishersHandler *handler.PublisherHandler
	AuthorsHandler    *handler.AuthorHandler
	TagsHandler       *handler.TagHandler
	BooksHandler      *handler.BookHandler
	UserHandler       *handler.UserHandler
	UserIDExtractor   middleware.UserIDExtractor
	UserRoleChecker   middleware.RoleChecker

	UserUsecase interface {
		CheckUserAnyRole(ctx context.Context, userID uuid.UUID, needRoleList []model.Role) error
	}
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

	rt.HandleFunc("/user", deps.UserHandler.GetUserInfo).Methods(http.MethodGet)
	rt.HandleFunc("/user/register/email", deps.UserHandler.RegisterUserByEmail).Methods(http.MethodPost)
	rt.HandleFunc("/user/token/email", deps.UserHandler.GetUserTokenByEmail).Methods(http.MethodPost)

	rt.HandleFunc("/publishers", deps.PublishersHandler.ListPublishers).Methods(http.MethodGet)
	rt.HandleFunc("/publishers/{id}", deps.PublishersHandler.GetPublisher).Methods(http.MethodGet)

	rt.HandleFunc("/authors", deps.AuthorsHandler.ListAuthors).Methods(http.MethodGet)
	rt.HandleFunc("/authors/{id}", deps.AuthorsHandler.GetAuthor).Methods(http.MethodGet)

	rt.HandleFunc("/tags", deps.TagsHandler.ListTags).Methods(http.MethodGet)
	rt.HandleFunc("/tags/{id}", deps.TagsHandler.GetTag).Methods(http.MethodGet)

	rt.HandleFunc("/books", deps.BooksHandler.ListBooks).Methods(http.MethodGet)
	rt.HandleFunc("/books/{id}", deps.BooksHandler.GetBook).Methods(http.MethodGet)

	write := rt.NewRoute().Subrouter()
	write.Use(middleware.RequireAuth(deps.UserIDExtractor))
	write.Use(middleware.RequireAnyRole(deps.UserRoleChecker, []model.Role{model.RoleAdmin}))

	write.HandleFunc("/publishers", deps.PublishersHandler.AddPublisher).Methods(http.MethodPost)
	write.HandleFunc("/publishers/{id}", deps.PublishersHandler.UpdatePublisher).Methods(http.MethodPut)
	write.HandleFunc("/publishers/{id}", deps.PublishersHandler.RemovePublisher).Methods(http.MethodDelete)

	write.HandleFunc("/authors", deps.AuthorsHandler.AddAuthor).Methods(http.MethodPost)
	write.HandleFunc("/authors/{id}", deps.AuthorsHandler.UpdateAuthor).Methods(http.MethodPut)
	write.HandleFunc("/authors/{id}", deps.AuthorsHandler.RemoveAuthor).Methods(http.MethodDelete)

	write.HandleFunc("/tags", deps.TagsHandler.AddTag).Methods(http.MethodPost)
	write.HandleFunc("/tags/{id}", deps.TagsHandler.UpdateTag).Methods(http.MethodPut)
	write.HandleFunc("/tags/{id}", deps.TagsHandler.RemoveTag).Methods(http.MethodDelete)

	write.HandleFunc("/books", deps.BooksHandler.AddBook).Methods(http.MethodPost)
	write.HandleFunc("/books/{id}", deps.BooksHandler.UpdateBook).Methods(http.MethodPut)
	write.HandleFunc("/books/{id}", deps.BooksHandler.RemoveBook).Methods(http.MethodDelete)

	return rt, nil
}

func HealthHandler(writer http.ResponseWriter, _ *http.Request) {
	writer.WriteHeader(http.StatusOK)
}
