package books

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/iamvkosarev/book-shelf/internal/model"
	"time"
)

type CreateBookInput struct {
	Title       string
	PublisherID *uuid.UUID
	TagsIDs     []uuid.UUID
	AuthorsIDs  []uuid.UUID
	PublishedAt *time.Time
	Description *string
	Price       *float64
}

type UpdateBookPatch struct {
	PublisherID *uuid.UUID
	TagsIDs     []uuid.UUID
	AuthorsIDs  []uuid.UUID
	PublishedAt *time.Time
	Title       *string
	Description *string
	Price       *float64
}

type ListBookParameters struct {
	AuthorsIDs []uuid.UUID
}

type BooksStorage interface {
	AddBook(ctx context.Context, input CreateBookInput) (uuid.UUID, error)
	GetBook(ctx context.Context, id uuid.UUID) (model.Book, error)
	UpdateBook(ctx context.Context, id uuid.UUID, patch UpdateBookPatch) error
	RemoveBook(ctx context.Context, id uuid.UUID) error
	ListBooks(ctx context.Context, parameters ListBookParameters) ([]model.Book, error)
}

type AuthorsUsecase interface {
	GetAuthor(ctx context.Context, id uuid.UUID, expandPerson bool) (model.Author, error)
}
type PublishersUsecase interface {
	GetPublisher(ctx context.Context, id uuid.UUID) (model.Publisher, error)
}
type TagsUsecase interface {
	GetTag(ctx context.Context, id uuid.UUID) (model.Tag, error)
}

type BooksUsecase struct {
	booksStorage      BooksStorage
	authorsUsecase    AuthorsUsecase
	publishersUsecase PublishersUsecase
	tagsUsecase       TagsUsecase
}

func NewBooksUsecase(
	booksStorage BooksStorage,
	authorsUsecase AuthorsUsecase,
	publishersUsecase PublishersUsecase,
	tagsUsecase TagsUsecase,
) *BooksUsecase {
	return &BooksUsecase{
		booksStorage:      booksStorage,
		authorsUsecase:    authorsUsecase,
		publishersUsecase: publishersUsecase,
		tagsUsecase:       tagsUsecase,
	}
}
func (p *BooksUsecase) AddBook(ctx context.Context, input CreateBookInput) (uuid.UUID, error) {
	if input.PublisherID != nil && *input.PublisherID != uuid.Nil {
		if _, err := p.publishersUsecase.GetPublisher(ctx, *input.PublisherID); err != nil {
			return uuid.Nil, fmt.Errorf("failed to validate publisher: %w", err)
		}
	}
	for _, authorID := range input.AuthorsIDs {
		if _, err := p.authorsUsecase.GetAuthor(ctx, authorID, false); err != nil {
			return uuid.Nil, fmt.Errorf("failed to validate author %s: %w", authorID, err)
		}
	}
	for _, tagID := range input.TagsIDs {
		if _, err := p.tagsUsecase.GetTag(ctx, tagID); err != nil {
			return uuid.Nil, fmt.Errorf("failed to validate tag %s: %w", tagID, err)
		}
	}

	id, err := p.booksStorage.AddBook(ctx, input)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to add book to storage: %w", err)
	}
	return id, nil
}

func (p *BooksUsecase) GetBook(
	ctx context.Context,
	id uuid.UUID,
	expandAuthors bool,
	expandTags bool,
	expandPublisher bool,
) (model.Book, error) {
	book, err := p.booksStorage.GetBook(ctx, id)
	if err != nil {
		return model.Book{}, fmt.Errorf("failed to get book from storage: %w", err)
	}

	if expandPublisher && book.PublisherID != nil && *book.PublisherID != uuid.Nil {
		publisher, err := p.publishersUsecase.GetPublisher(ctx, *book.PublisherID)
		if err != nil {
			return model.Book{}, fmt.Errorf("failed to expand publisher: %w", err)
		}
		book.Publisher = &publisher
	}

	if expandAuthors && len(book.AuthorsIDs) > 0 {
		authors := make([]model.Author, 0, len(book.AuthorsIDs))
		for _, aID := range book.AuthorsIDs {
			author, err := p.authorsUsecase.GetAuthor(ctx, aID, true)
			if err != nil {
				return model.Book{}, fmt.Errorf("failed to expand author %s: %w", aID, err)
			}
			authors = append(authors, author)
		}
		book.Authors = authors
	}

	if expandTags && len(book.TagsIDs) > 0 {
		tags := make([]model.Tag, 0, len(book.TagsIDs))
		for _, tID := range book.TagsIDs {
			tag, err := p.tagsUsecase.GetTag(ctx, tID)
			if err != nil {
				return model.Book{}, fmt.Errorf("failed to expand tag %s: %w", tID, err)
			}
			tags = append(tags, tag)
		}
		book.Tags = tags
	}

	return book, nil
}

func (p *BooksUsecase) UpdateBook(ctx context.Context, id uuid.UUID, patch UpdateBookPatch) error {
	if patch.PublisherID != nil && *patch.PublisherID != uuid.Nil {
		if _, err := p.publishersUsecase.GetPublisher(ctx, *patch.PublisherID); err != nil {
			return fmt.Errorf("failed to validate publisher: %w", err)
		}
	}
	if patch.AuthorsIDs != nil {
		for _, aID := range patch.AuthorsIDs {
			if _, err := p.authorsUsecase.GetAuthor(ctx, aID, false); err != nil {
				return fmt.Errorf("failed to validate author %s: %w", aID, err)
			}
		}
	}
	if patch.TagsIDs != nil {
		for _, tID := range patch.TagsIDs {
			if _, err := p.tagsUsecase.GetTag(ctx, tID); err != nil {
				return fmt.Errorf("failed to validate tag %s: %w", tID, err)
			}
		}
	}

	if err := p.booksStorage.UpdateBook(ctx, id, patch); err != nil {
		return fmt.Errorf("failed to update book in storage: %w", err)
	}
	return nil
}

func (p *BooksUsecase) RemoveBook(ctx context.Context, id uuid.UUID) error {
	if err := p.booksStorage.RemoveBook(ctx, id); err != nil {
		return fmt.Errorf("failed to remove book from storage: %w", err)
	}
	return nil
}

func (p *BooksUsecase) ListBooks(
	ctx context.Context,
	parameters ListBookParameters,
	expandAuthors, expandTags, expandPublisher bool,
) ([]model.Book, error) {
	books, err := p.booksStorage.ListBooks(ctx, parameters)
	if err != nil {
		return nil, fmt.Errorf("failed to list books from storage: %w", err)
	}

	if !expandAuthors && !expandTags && !expandPublisher {
		return books, nil
	}

	for i, book := range books {
		if expandPublisher && book.PublisherID != nil && *book.PublisherID != uuid.Nil {
			publisher, err := p.publishersUsecase.GetPublisher(ctx, *book.PublisherID)
			if err != nil {
				return nil, fmt.Errorf("failed to expand publisher: %w", err)
			}
			book.Publisher = &publisher
		}

		if expandAuthors && len(book.AuthorsIDs) > 0 {
			authors := make([]model.Author, 0, len(book.AuthorsIDs))
			for _, aID := range book.AuthorsIDs {
				author, err := p.authorsUsecase.GetAuthor(ctx, aID, true)
				if err != nil {
					return nil, fmt.Errorf("failed to expand author %s: %w", aID, err)
				}
				authors = append(authors, author)
			}
			book.Authors = authors
		}

		if expandTags && len(book.TagsIDs) > 0 {
			tags := make([]model.Tag, 0, len(book.TagsIDs))
			for _, tID := range book.TagsIDs {
				tag, err := p.tagsUsecase.GetTag(ctx, tID)
				if err != nil {
					return nil, fmt.Errorf("failed to expand tag %s: %w", tID, err)
				}
				tags = append(tags, tag)
			}
			book.Tags = tags
		}

		books[i] = book
	}

	return books, nil
}
