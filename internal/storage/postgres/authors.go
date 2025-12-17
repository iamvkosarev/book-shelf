package postgres

import (
	"context"
	"errors"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/iamvkosarev/book-shelf/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
)

const (
	tableAuthors = "authors"

	columnPersonID  = "person_id"
	columnPseudonym = "pseudonym"
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

func (p *AuthorsStorage) AddAuthor(ctx context.Context, personID uuid.UUID, pseudonym string) (uuid.UUID, error) {
	if personID == uuid.Nil && strings.TrimSpace(pseudonym) == "" {
		return uuid.Nil, model.ErrAuthorInvalidFields
	}
	sql, args, err := p.psql.Insert(tableAuthors).Columns(
		columnPersonID, columnPseudonym,
	).Values(toPostgresUUID(personID), toPostgresText(pseudonym)).Suffix(
		"RETURNING " + columnID,
	).ToSql()
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
			}
		}
		return uuid.Nil, err
	}
	return id, nil
}

func (p *AuthorsStorage) GetAuthor(ctx context.Context, id uuid.UUID) (model.Author, error) {
	sql, args, err := p.psql.Select(
		columnID, columnPersonID, columnPseudonym,
	).From(tableAuthors).Where(squirrel.Eq{columnID: id}).ToSql()
	if err != nil {
		return model.Author{}, err
	}

	var (
		author    model.Author
		personID  pgtype.UUID
		pseudonym pgtype.Text
	)

	if err = p.pool.QueryRow(ctx, sql, args...).Scan(
		&author.ID, &personID, &pseudonym,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Author{}, model.ErrAuthorNotFound
		}
		return model.Author{}, err
	}
	if personID.Valid {
		author.PersonID = personID.Bytes
	} else {
		author.PersonID = uuid.Nil
	}
	if pseudonym.Valid {
		author.Pseudonym = pseudonym.String
	} else {
		author.Pseudonym = ""
	}
	return author, nil
}

func (p *AuthorsStorage) UpdateAuthor(ctx context.Context, id uuid.UUID, author model.Author) error {
	sql, args, err := p.psql.Update(tableAuthors).
		Set(columnPersonID, toPostgresUUID(author.PersonID)).
		Set(columnPseudonym, toPostgresText(author.Pseudonym)).
		Where(squirrel.Eq{columnID: id}).
		ToSql()
	if err != nil {
		return err
	}

	commandTag, err := p.pool.Exec(ctx, sql, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.ErrAuthorAlreadyExists
		}
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return model.ErrAuthorNotFound
	}
	return nil
}

func (p *AuthorsStorage) RemoveAuthor(ctx context.Context, id uuid.UUID) error {
	sql, args, err := p.psql.Delete(tableAuthors).Where(squirrel.Eq{columnID: id}).ToSql()
	if err != nil {
		return err
	}
	if _, err = p.pool.Exec(ctx, sql, args...); err != nil {
		return err
	}
	return nil
}

func (p *AuthorsStorage) ListAuthors(ctx context.Context) ([]model.Author, error) {
	sql, args, err := p.psql.Select(columnID, columnPersonID, columnPseudonym).
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
			a      model.Author
			person pgtype.UUID
			pseudo pgtype.Text
		)

		if err = rows.Scan(&a.ID, &person, &pseudo); err != nil {
			return nil, err
		}

		if person.Valid {
			a.PersonID = person.Bytes
		} else {
			a.PersonID = uuid.Nil
		}

		if pseudo.Valid {
			a.Pseudonym = pseudo.String
		} else {
			a.Pseudonym = ""
		}

		authors = append(authors, a)
	}
	return authors, rows.Err()
}
