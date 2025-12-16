package middleware

import (
	"github.com/iamvkosarev/book-shelf/pkg/logs"
	"net/http"
)

func HandlePanic(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					logs.Panic(r.Context(), "catch panic in middleware", err)
				}
			}()
			h.ServeHTTP(w, r)
		},
	)
}
