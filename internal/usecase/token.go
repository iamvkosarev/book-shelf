package usecase

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/golang-jwt/jwt/v5/request"
	"github.com/google/uuid"
	"github.com/iamvkosarev/book-shelf/config"
	"github.com/iamvkosarev/book-shelf/internal/model"
	"net/http"
	"strings"
	"time"
)

type Claims struct {
	jwt.RegisteredClaims
	UserId uuid.UUID `json:"user_id"`
}

type TokenUsecase struct {
	config    config.Authorization
	verifyKey *rsa.PublicKey
	signKey   *rsa.PrivateKey
}

func NewTokenUsecase(cfg config.Authorization) (*TokenUsecase, error) {
	privateKey := strings.ReplaceAll(cfg.PrivateKey, `\n`, "\n")
	publicKey := strings.ReplaceAll(cfg.PublicKey, `\n`, "\n")
	signKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKey))
	if err != nil {
		return nil, err
	}

	verifyKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(publicKey))
	if err != nil {
		return nil, err
	}

	return &TokenUsecase{
		config:    cfg,
		signKey:   signKey,
		verifyKey: verifyKey,
	}, nil
}

func (u *TokenUsecase) GetVerifiedUserIDFromRequest(r *http.Request) (uuid.UUID, error) {
	token, err := request.ParseFromRequest(
		r, request.OAuth2Extractor, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok || t.Method.Alg() != jwt.SigningMethodRS256.Alg() {
				return nil, fmt.Errorf("unexpected signing method: %s", t.Method.Alg())
			}
			return u.verifyKey, nil
		}, request.WithClaims(&Claims{}),
	)
	if err != nil {
		switch {
		case errors.Is(err, request.ErrNoTokenInRequest):
			return uuid.Nil, model.ErrTokenNotFound
		case errors.Is(err, jwt.ErrSignatureInvalid):
			return uuid.Nil, model.ErrSignatureInvalid
		case errors.Is(err, jwt.ErrTokenExpired):
			return uuid.Nil, model.ErrTokenExpired
		case errors.Is(err, rsa.ErrVerification):
			return uuid.Nil, model.ErrTokenVerification
		case errors.Is(err, rsa.ErrDecryption):
			return uuid.Nil, model.ErrTokenDecryption
		default:
			return uuid.Nil, err
		}
	}

	cls, ok := token.Claims.(*Claims)
	if !ok {
		return uuid.Nil, model.ErrParseClaims
	}
	return cls.UserId, nil
}

func (u *TokenUsecase) GenerateUserToken(userID uuid.UUID) (string, error) {
	t := jwt.NewWithClaims(
		jwt.SigningMethodRS256,
		&Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(u.config.TokenTTL)),
			},
			UserId: userID,
		},
	)
	tokenString, err := t.SignedString(u.signKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return tokenString, nil
}
