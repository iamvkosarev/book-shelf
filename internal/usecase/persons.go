package usecase

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/iamvkosarev/book-shelf/internal/model"
)

type PersonsStorage interface {
	AddPerson(ctx context.Context, firstName, lastName, middleName string) (uuid.UUID, error)
	GetPerson(ctx context.Context, id uuid.UUID) (model.Person, error)
	UpdatePerson(ctx context.Context, id uuid.UUID, person model.Person) error
	RemovePerson(ctx context.Context, id uuid.UUID) error
	ListPersons(ctx context.Context) ([]model.Person, error)
}

type PersonsUsecase struct {
	storage PersonsStorage
}

func NewPersonsUsecase(storage PersonsStorage) *PersonsUsecase {
	return &PersonsUsecase{
		storage: storage,
	}
}

func (p *PersonsUsecase) AddPerson(ctx context.Context, firstName, lastName, middleName string) (uuid.UUID, error) {
	id, err := p.storage.AddPerson(ctx, firstName, lastName, middleName)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to add person to storage: %w", err)
	}
	return id, nil
}

func (p *PersonsUsecase) GetPerson(ctx context.Context, id uuid.UUID) (model.Person, error) {
	person, err := p.storage.GetPerson(ctx, id)
	if err != nil {
		return model.Person{}, fmt.Errorf("failed to get person from storage: %w", err)
	}
	return person, nil
}

func (p *PersonsUsecase) UpdatePerson(ctx context.Context, id uuid.UUID, firstName, lastName, middleName string) error {
	person, err := p.storage.GetPerson(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get person from storage: %w", err)
	}
	if firstName != "" {
		person.FirstName = firstName
	}
	if lastName != "" {
		person.LastName = lastName
	}
	if middleName != "" {
		person.MiddleName = middleName
	}
	err = p.storage.UpdatePerson(ctx, id, person)
	if err != nil {
		return fmt.Errorf("failed to update person in storage: %w", err)
	}
	return nil
}

func (p *PersonsUsecase) RemovePerson(ctx context.Context, id uuid.UUID) error {
	err := p.storage.RemovePerson(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to remove person from storage: %w", err)
	}
	return nil
}

func (p *PersonsUsecase) ListPersons(ctx context.Context) ([]model.Person, error) {
	persons, err := p.storage.ListPersons(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get list Persons from storage: %w", err)
	}
	return persons, nil
}
