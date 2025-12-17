package usecase

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/iamvkosarev/book-shelf/internal/model"
)

type AuthorsStorage interface {
	AddAuthor(ctx context.Context, personID uuid.UUID, pseudonym string) (uuid.UUID, error)
	GetAuthor(ctx context.Context, id uuid.UUID) (model.Author, error)
	UpdateAuthor(ctx context.Context, id uuid.UUID, author model.Author) error
	RemoveAuthor(ctx context.Context, id uuid.UUID) error
	ListAuthors(ctx context.Context) ([]model.Author, error)
}

type AuthorsUsecase struct {
	storage       AuthorsStorage
	personUsecase *PersonsUsecase
}

func NewAuthorsUsecase(storage AuthorsStorage, personUsecase *PersonsUsecase) *AuthorsUsecase {
	return &AuthorsUsecase{
		storage:       storage,
		personUsecase: personUsecase,
	}
}

func (p *AuthorsUsecase) AddAuthor(
	ctx context.Context, personID uuid.UUID, pseudonym string,
) (uuid.UUID, error) {
	_, err := p.personUsecase.GetPerson(ctx, personID)
	if err != nil && !errors.Is(err, model.ErrPersonNotFound) {
		return uuid.Nil, err
	}
	if errors.Is(err, model.ErrPersonNotFound) && pseudonym == "" {
		return uuid.Nil, model.ErrAuthorInvalidFields
	}
	id, err := p.storage.AddAuthor(ctx, personID, pseudonym)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to add author to storage: %w", err)
	}
	return id, nil
}

func (p *AuthorsUsecase) GetAuthor(ctx context.Context, id uuid.UUID, expendPersonData bool) (model.Author, error) {
	author, err := p.storage.GetAuthor(ctx, id)
	if err != nil {
		return model.Author{}, fmt.Errorf("failed to get author from storage: %w", err)
	}
	if expendPersonData && author.PersonID != uuid.Nil {
		person, err := p.personUsecase.GetPerson(ctx, author.PersonID)
		if err != nil {
			return model.Author{}, fmt.Errorf("failed to get person from storage: id %s; %w", author.PersonID, err)
		}
		author.Person = person
	}
	return author, nil
}

func (p *AuthorsUsecase) UpdateAuthor(ctx context.Context, id, personID uuid.UUID, pseudonym string) error {
	author, err := p.storage.GetAuthor(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get author from storage: %w", err)
	}
	if pseudonym != "" {
		author.Pseudonym = pseudonym
	}
	if personID != uuid.Nil {
		author.PersonID = personID
	}
	err = p.storage.UpdateAuthor(ctx, id, author)
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

func (p *AuthorsUsecase) ListAuthors(ctx context.Context, expendPersonData bool) ([]model.Author, error) {
	authors, err := p.storage.ListAuthors(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get list Authors from storage: %w", err)
	}
	if expendPersonData {
		for i, author := range authors {
			if author.PersonID != uuid.Nil {
				person, err := p.personUsecase.GetPerson(ctx, author.PersonID)
				if err != nil {
					return nil, fmt.Errorf("failed to get person from storage: id %s; %w", author.PersonID, err)
				}
				authors[i].Person = person
			}
		}
	}
	return authors, nil
}
