package usecase

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/iamvkosarev/book-shelf/internal/model"
)

type GenresStorage interface {
	AddGenre(ctx context.Context, name string) (uuid.UUID, error)
	GetGenre(ctx context.Context, id uuid.UUID) (model.Genre, error)
	UpdateGenre(ctx context.Context, id uuid.UUID, genre model.Genre) error
	RemoveGenre(ctx context.Context, id uuid.UUID) error
	ListGenres(ctx context.Context) ([]model.Genre, error)
}

type GenresUsecase struct {
	storage GenresStorage
}

func NewGenresUsecase(storage GenresStorage) *GenresUsecase {
	return &GenresUsecase{
		storage: storage,
	}
}

func (p *GenresUsecase) AddGenre(ctx context.Context, name string) (uuid.UUID, error) {
	id, err := p.storage.AddGenre(ctx, name)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to add genre to storage: %w", err)
	}
	return id, nil
}

func (p *GenresUsecase) GetGenre(ctx context.Context, id uuid.UUID) (model.Genre, error) {
	genre, err := p.storage.GetGenre(ctx, id)
	if err != nil {
		return model.Genre{}, fmt.Errorf("failed to get genre from storage: %w", err)
	}
	return genre, nil
}

func (p *GenresUsecase) UpdateGenre(ctx context.Context, id uuid.UUID, name string) error {
	genre, err := p.storage.GetGenre(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get genre from storage: %w", err)
	}
	genre.Name = name
	err = p.storage.UpdateGenre(ctx, id, genre)
	if err != nil {
		return fmt.Errorf("failed to update genre in storage: %w", err)
	}
	return nil
}

func (p *GenresUsecase) RemoveGenre(ctx context.Context, id uuid.UUID) error {
	err := p.storage.RemoveGenre(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to remove genre from storage: %w", err)
	}
	return nil
}

func (p *GenresUsecase) ListGenres(ctx context.Context) ([]model.Genre, error) {
	genres, err := p.storage.ListGenres(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get list genres from storage: %w", err)
	}
	return genres, nil
}
