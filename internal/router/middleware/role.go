package middleware

import (
	"context"
	"github.com/iamvkosarev/book-shelf/internal/handler"
	"net/http"

	"github.com/google/uuid"
	"github.com/iamvkosarev/book-shelf/internal/model"
)

type RoleChecker interface {
	CheckUserAnyRole(ctx context.Context, userID uuid.UUID, needRoleList []model.Role) error
}

func RequireAnyRole(checker RoleChecker, roles []model.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(writer http.ResponseWriter, request *http.Request) {
				userID, ok := UserIDFromContext(request.Context())
				if !ok {
					handler.SendError(writer, model.ErrTokenNotFound)
					return
				}

				if err := checker.CheckUserAnyRole(request.Context(), userID, roles); err != nil {
					handler.SendError(writer, err)
					return
				}

				next.ServeHTTP(writer, request)
			},
		)
	}
}
