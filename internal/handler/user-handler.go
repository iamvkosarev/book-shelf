package handler

import (
	"context"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/iamvkosarev/book-shelf/internal/model"
	"github.com/iamvkosarev/book-shelf/pkg/logs"
	"net/http"
)

type UserUsecase interface {
	AuthenticateByEmail(ctx context.Context, email, password string) (string, error)
	RegisterByEmail(ctx context.Context, email, password string) (uuid.UUID, error)
	GetUserInfo(ctx context.Context, userID uuid.UUID) (*model.User, error)
}

type UserIDExtractor interface {
	GetVerifiedUserIDFromRequest(r *http.Request) (uuid.UUID, error)
}

type UserHandler struct {
	userUsecase     UserUsecase
	userIDExtractor UserIDExtractor
	validate        *validator.Validate
}

func NewUserHandler(userUsecase UserUsecase, userIDExtractor UserIDExtractor) *UserHandler {
	return &UserHandler{
		userUsecase:     userUsecase,
		userIDExtractor: userIDExtractor,
		validate:        validator.New(),
	}
}

type GetUserInfoResponse struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

func (h *UserHandler) GetUserInfo(writer http.ResponseWriter, request *http.Request) {
	userID, err := h.userIDExtractor.GetVerifiedUserIDFromRequest(request)
	if err != nil {
		SendError(writer, err)
		logs.Error("failed to get user id", err)
		return
	}
	user, err := h.userUsecase.GetUserInfo(request.Context(), userID)
	if err != nil {
		SendError(writer, err)
		logs.Error("failed to get user info", err)
		return
	}
	response := GetUserInfoResponse{
		ID:    user.ID,
		Email: user.Email,
	}
	sendOkJSON(writer, response)
}

type RegisterUserRequest struct {
	Email    string `json:"email" validate:"required,email" example:"test@test.ru"`
	Password string `json:"password" validate:"required,min=8" example:"TestTest123"`
}

func (h *UserHandler) RegisterUserByEmail(writer http.ResponseWriter, request *http.Request) {
	var requestData RegisterUserRequest
	if ok := decode(writer, request, &requestData); !ok {
		return
	}
	if validationErr(writer, h.validate, requestData) {
		return
	}
	id, err := h.userUsecase.RegisterByEmail(request.Context(), requestData.Email, requestData.Password)
	if err != nil {
		SendError(writer, err)
		logs.Error("failed to register user by tag", err)
		return
	}
	sendCreatedJSON(writer, AddTagResponse{ID: id})
}

type AuthenticateUserRequest struct {
	Email    string `json:"email" validate:"required,email" example:"test@test.ru"`
	Password string `json:"password" validate:"required,min=8" example:"TestTest123"`
}

type AuthenticateUserResponse struct {
	Token string `json:"token"`
}

func (h *UserHandler) GetUserTokenByEmail(writer http.ResponseWriter, request *http.Request) {
	var requestData AuthenticateUserRequest
	if ok := decode(writer, request, &requestData); !ok {
		return
	}
	if validationErr(writer, h.validate, requestData) {
		return
	}
	token, err := h.userUsecase.AuthenticateByEmail(request.Context(), requestData.Email, requestData.Password)
	if err != nil {
		SendError(writer, err)
		logs.Error("failed to authenticate the user", err)
		return
	}
	response := AuthenticateUserResponse{
		Token: token,
	}
	sendOkJSON(writer, response)
}
