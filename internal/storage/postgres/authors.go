package postgres

import (
	"context"
	"errors"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/iamvkosarev/book-shelf/internal/model"
	"github.com/iamvkosarev/book-shelf/internal/usecase"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	tableAuthors = "authors"

	columnPseudonym  = "pseudonym"
	columnFirstName  = "first_name"
	columnLastName   = "last_name"
	columnMiddleName = "middle_name"
)

type AuthorsStorage struct {
	pool *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewAuthorsStorage(pool *pgxpool.Pool) *AuthorsStorage {
	return &AuthorsStorage{
		pool: pool,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (p *AuthorsStorage) AddAuthor(ctx context.Context, input usecase.AddAuthorInput) (uuid.UUID, error) {
	if authorIdentityCount(input.FirstName, input.LastName, input.Pseudonym) != 1 {
		return uuid.Nil, model.ErrAuthorInvalidFields
	}

	firstName := toNonEmptyTrimmedPtr(input.FirstName)
	lastName := toNonEmptyTrimmedPtr(input.LastName)
	pseudonym := toNonEmptyTrimmedPtr(input.Pseudonym)

	var middleName any = nil
	if input.MiddleName != nil {
		middleName = strPrtToAny(input.MiddleName)
	}

	sql, args, err := p.psql.
		Insert(tableAuthors).
		Columns(columnFirstName, columnLastName, columnMiddleName, columnPseudonym).
		Values(firstName, lastName, middleName, pseudonym).
		Suffix("RETURNING " + columnID).
		ToSql()
	if err != nil {
		return uuid.Nil, err
	}

	var id uuid.UUID
	if err = p.pool.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return uuid.Nil, model.ErrAuthorAlreadyExists
			case "23514":
				return uuid.Nil, model.ErrAuthorInvalidFields
			}
		}
		return uuid.Nil, err
	}

	return id, nil
}

func (p *AuthorsStorage) GetAuthor(ctx context.Context, id uuid.UUID) (model.Author, error) {
	sql, args, err := p.psql.
		Select(columnID, columnFirstName, columnLastName, columnMiddleName, columnPseudonym).
		From(tableAuthors).
		Where(squirrel.Eq{columnID: id}).
		ToSql()
	if err != nil {
		return model.Author{}, err
	}

	var (
		a                                          model.Author
		firstName, lastName, middleName, pseudonym pgtype.Text
	)

	if err = p.pool.QueryRow(ctx, sql, args...).Scan(
		&a.ID, &firstName, &lastName, &middleName, &pseudonym,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Author{}, model.ErrAuthorNotFound
		}
		return model.Author{}, err
	}

	a.FirstName = postgresTextToStrPtr(firstName)
	a.LastName = postgresTextToStrPtr(lastName)
	a.MiddleName = postgresTextToStrPtr(middleName)
	a.Pseudonym = postgresTextToStrPtr(pseudonym)

	return a, nil
}

func (p *AuthorsStorage) UpdateAuthor(ctx context.Context, id uuid.UUID, input usecase.UpdateAuthorInput) error {
	q := p.psql.Update(tableAuthors).Where(squirrel.Eq{columnID: id})

	sets := 0

	if input.FirstName != nil {
		q = q.Set(columnFirstName, strPrtToAny(input.FirstName))
		sets++
	}
	if input.LastName != nil {
		q = q.Set(columnLastName, strPrtToAny(input.LastName))
		sets++
	}
	if input.Pseudonym != nil {
		q = q.Set(columnPseudonym, strPrtToAny(input.Pseudonym))
		sets++
	}
	if input.MiddleName != nil {
		q = q.Set(columnMiddleName, strPrtToAny(input.MiddleName))
		sets++
	}

	if sets == 0 {
		return nil
	}

	sql, args, err := q.ToSql()
	if err != nil {
		return err
	}

	tag, err := p.pool.Exec(ctx, sql, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return model.ErrAuthorAlreadyExists
			case "23514":
				return model.ErrAuthorInvalidFields
			}
		}
		return err
	}

	if tag.RowsAffected() == 0 {
		return model.ErrAuthorNotFound
	}

	return nil
}

func (p *AuthorsStorage) RemoveAuthor(ctx context.Context, id uuid.UUID) error {
	sql, args, err := p.psql.
		Delete(tableAuthors).
		Where(squirrel.Eq{columnID: id}).
		ToSql()
	if err != nil {
		return err
	}

	tag, err := p.pool.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return model.ErrAuthorNotFound
	}

	return nil
}

func (p *AuthorsStorage) ListAuthors(ctx context.Context) ([]model.Author, error) {
	sql, args, err := p.psql.
		Select(columnID, columnFirstName, columnLastName, columnMiddleName, columnPseudonym).
		From(tableAuthors).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := p.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var authors []model.Author
	for rows.Next() {
		var (
			a                                          model.Author
			firstName, lastName, middleName, pseudonym pgtype.Text
		)

		if err = rows.Scan(&a.ID, &firstName, &lastName, &middleName, &pseudonym); err != nil {
			return nil, err
		}

		a.FirstName = postgresTextToStrPtr(firstName)
		a.LastName = postgresTextToStrPtr(lastName)
		a.MiddleName = postgresTextToStrPtr(middleName)
		a.Pseudonym = postgresTextToStrPtr(pseudonym)

		authors = append(authors, a)
	}

	return authors, rows.Err()
}

func authorIdentityCount(firstName, lastName, pseudonym *string) int {
	count := 0
	if toNonEmptyTrimmedPtr(firstName) != nil {
		count++
	}
	if toNonEmptyTrimmedPtr(lastName) != nil {
		count++
	}
	if toNonEmptyTrimmedPtr(pseudonym) != nil {
		count++
	}
	return count
}
