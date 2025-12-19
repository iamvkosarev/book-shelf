package usecase

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/iamvkosarev/book-shelf/internal/model"
)

type AddAuthorInput struct {
	FirstName  *string
	LastName   *string
	MiddleName *string
	Pseudonym  *string
}

type UpdateAuthorInput struct {
	FirstName  *string
	LastName   *string
	MiddleName *string
	Pseudonym  *string
}

type AuthorsStorage interface {
	AddAuthor(ctx context.Context, input AddAuthorInput) (uuid.UUID, error)
	GetAuthor(ctx context.Context, id uuid.UUID) (model.Author, error)
	UpdateAuthor(ctx context.Context, id uuid.UUID, input UpdateAuthorInput) error
	RemoveAuthor(ctx context.Context, id uuid.UUID) error
	ListAuthors(ctx context.Context) ([]model.Author, error)
}

type AuthorsUsecase struct {
	storage AuthorsStorage
}

func NewAuthorsUsecase(storage AuthorsStorage) *AuthorsUsecase {
	return &AuthorsUsecase{
		storage: storage,
	}
}

func (p *AuthorsUsecase) AddAuthor(
	ctx context.Context, input AddAuthorInput,
) (uuid.UUID, error) {
	if !hasIdentity(input.FirstName, input.LastName, input.Pseudonym) {
		return uuid.Nil, model.ErrAuthorInvalidFields
	}
	id, err := p.storage.AddAuthor(ctx, input)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to add author to storage: %w", err)
	}
	return id, nil
}

func (p *AuthorsUsecase) GetAuthor(ctx context.Context, id uuid.UUID) (model.Author, error) {
	author, err := p.storage.GetAuthor(ctx, id)
	if err != nil {
		return model.Author{}, fmt.Errorf("failed to get author from storage: %w", err)
	}
	return author, nil
}

func (p *AuthorsUsecase) UpdateAuthor(ctx context.Context, id uuid.UUID, input UpdateAuthorInput) error {
	author, err := p.storage.GetAuthor(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get author from storage: %w", err)
	}
	if input.FirstName != nil {
		author.FirstName = input.FirstName
	}
	if input.LastName != nil {
		author.LastName = input.LastName
	}
	if input.Pseudonym != nil {
		author.Pseudonym = input.Pseudonym
	}

	if !hasIdentity(author.FirstName, author.LastName, author.Pseudonym) {
		return model.ErrAuthorInvalidFields
	}

	err = p.storage.UpdateAuthor(ctx, id, input)
	if err != nil {
		return fmt.Errorf("failed to update author in storage: %w", err)
	}
	return nil
}

func (p *AuthorsUsecase) RemoveAuthor(ctx context.Context, id uuid.UUID) error {
	err := p.storage.RemoveAuthor(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to remove author from storage: %w", err)
	}
	return nil
}

func (p *AuthorsUsecase) ListAuthors(ctx context.Context) ([]model.Author, error) {
	authors, err := p.storage.ListAuthors(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get list Authors from storage: %w", err)
	}
	return authors, nil
}

func hasIdentity(first, last, pseudonym *string) bool {
	return (first != nil && *first != "") ||
		(last != nil && *last != "") ||
		(pseudonym != nil && *pseudonym != "")
}
