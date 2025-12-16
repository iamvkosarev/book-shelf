package router

import (
	"github.com/gorilla/mux"
	"github.com/iamvkosarev/book-shelf/config"
	"github.com/iamvkosarev/book-shelf/internal/router/middleware"
	"net/http"
)

type Deps struct {
}

func Setup(rt *mux.Router, deps Deps, cfg config.Router) (http.Handler, error) {
	rt.Use(middleware.HandlePanic)
	rt.Use(
		func(next http.Handler) http.Handler {
			return http.TimeoutHandler(next, cfg.APITimeout, "request timeout")
		},
	)
	rt.Use(middleware.LogProcessesEdges)
	return rt, nil
}
