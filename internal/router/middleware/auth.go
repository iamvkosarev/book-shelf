package middleware

import (
	"context"
	"github.com/iamvkosarev/book-shelf/internal/handler"
	"net/http"

	"github.com/google/uuid"
)

type UserIDExtractor interface {
	GetVerifiedUserIDFromRequest(r *http.Request) (uuid.UUID, error)
}

type ctxKey int

const userIDKey ctxKey = iota

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	v := ctx.Value(userIDKey)
	id, ok := v.(uuid.UUID)
	return id, ok
}

func RequireAuth(extractor UserIDExtractor) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(writer http.ResponseWriter, request *http.Request) {
				userID, err := extractor.GetVerifiedUserIDFromRequest(request)
				if err != nil {
					handler.SendError(writer, err)
					return
				}

				ctx := context.WithValue(request.Context(), userIDKey, userID)
				next.ServeHTTP(writer, request.WithContext(ctx))
			},
		)
	}
}
