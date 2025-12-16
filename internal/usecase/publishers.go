package usecase

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/iamvkosarev/book-shelf/internal/model"
)

type PublishersStorage interface {
	AddPublisher(ctx context.Context, name string) (uuid.UUID, error)
	GetPublisher(ctx context.Context, id uuid.UUID) (model.Publisher, error)
	UpdatePublisher(ctx context.Context, id uuid.UUID, publisher model.Publisher) error
	RemovePublisher(ctx context.Context, id uuid.UUID) error
	ListPublishers(ctx context.Context) ([]model.Publisher, error)
}

type PublishersUsecase struct {
	storage PublishersStorage
}

func NewPublishersUsecase(storage PublishersStorage) *PublishersUsecase {
	return &PublishersUsecase{
		storage: storage,
	}
}

func (p *PublishersUsecase) AddPublisher(ctx context.Context, name string) (uuid.UUID, error) {
	id, err := p.storage.AddPublisher(ctx, name)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to add publisher to storage: %w", err)
	}
	return id, nil
}

func (p *PublishersUsecase) GetPublisher(ctx context.Context, id uuid.UUID) (model.Publisher, error) {
	publisher, err := p.storage.GetPublisher(ctx, id)
	if err != nil {
		return model.Publisher{}, fmt.Errorf("failed to get publisher from storage: %w", err)
	}
	return publisher, nil
}

func (p *PublishersUsecase) UpdatePublisher(ctx context.Context, id uuid.UUID, name string) error {
	publisher, err := p.storage.GetPublisher(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get publisher from storage: %w", err)
	}
	publisher.Name = name
	err = p.storage.UpdatePublisher(ctx, id, publisher)
	if err != nil {
		return fmt.Errorf("failed to update publisher in storage: %w", err)
	}
	return nil
}

func (p *PublishersUsecase) RemovePublisher(ctx context.Context, id uuid.UUID) error {
	err := p.storage.RemovePublisher(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to remove publisher from storage: %w", err)
	}
	return nil
}

func (p *PublishersUsecase) ListPublishers(ctx context.Context) ([]model.Publisher, error) {
	publishers, err := p.storage.ListPublishers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get list publishers from storage: %w", err)
	}
	return publishers, nil
}
