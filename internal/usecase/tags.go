package usecase

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/iamvkosarev/book-shelf/internal/model"
)

type TagsStorage interface {
	AddTag(ctx context.Context, name string) (uuid.UUID, error)
	GetTag(ctx context.Context, id uuid.UUID) (model.Tag, error)
	UpdateTag(ctx context.Context, id uuid.UUID, tag model.Tag) error
	RemoveTag(ctx context.Context, id uuid.UUID) error
	ListTags(ctx context.Context) ([]model.Tag, error)
}

type TagsUsecase struct {
	storage TagsStorage
}

func NewTagsUsecase(storage TagsStorage) *TagsUsecase {
	return &TagsUsecase{
		storage: storage,
	}
}

func (p *TagsUsecase) AddTag(ctx context.Context, name string) (uuid.UUID, error) {
	id, err := p.storage.AddTag(ctx, name)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to add tag to storage: %w", err)
	}
	return id, nil
}

func (p *TagsUsecase) GetTag(ctx context.Context, id uuid.UUID) (model.Tag, error) {
	tag, err := p.storage.GetTag(ctx, id)
	if err != nil {
		return model.Tag{}, fmt.Errorf("failed to get tag from storage: %w", err)
	}
	return tag, nil
}

func (p *TagsUsecase) UpdateTag(ctx context.Context, id uuid.UUID, name string) error {
	tag, err := p.storage.GetTag(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get tag from storage: %w", err)
	}
	tag.Name = name
	err = p.storage.UpdateTag(ctx, id, tag)
	if err != nil {
		return fmt.Errorf("failed to update tag in storage: %w", err)
	}
	return nil
}

func (p *TagsUsecase) RemoveTag(ctx context.Context, id uuid.UUID) error {
	err := p.storage.RemoveTag(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to remove tag from storage: %w", err)
	}
	return nil
}

func (p *TagsUsecase) ListTags(ctx context.Context) ([]model.Tag, error) {
	tags, err := p.storage.ListTags(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get list tags from storage: %w", err)
	}
	return tags, nil
}
