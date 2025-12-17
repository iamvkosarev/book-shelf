package postgres

import (
	"context"
	"errors"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/iamvkosarev/book-shelf/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tableGenres = "genres"

type GenresStorage struct {
	pool *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewGenresStorage(pool *pgxpool.Pool) *GenresStorage {
	return &GenresStorage{
		pool: pool,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (p *GenresStorage) AddGenre(ctx context.Context, name string) (uuid.UUID, error) {
	sql, args, err := p.psql.Insert(tableGenres).Columns(columnName).Values(name).Suffix(
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
				return uuid.Nil, model.ErrGenreAlreadyExists
			}
		}
		return uuid.Nil, err
	}
	return id, nil
}

func (p *GenresStorage) GetGenre(ctx context.Context, id uuid.UUID) (model.Genre, error) {
	sql, args, err := p.psql.Select(
		columnID, columnName,
	).From(tableGenres).Where(squirrel.Eq{columnID: id}).ToSql()
	if err != nil {
		return model.Genre{}, err
	}
	var genre model.Genre
	if err = p.pool.QueryRow(ctx, sql, args...).Scan(&genre.ID, &genre.Name); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Genre{}, model.ErrGenreNotFound
		}
		return model.Genre{}, err
	}
	return genre, nil
}

func (p *GenresStorage) UpdateGenre(ctx context.Context, id uuid.UUID, genre model.Genre) error {
	sql, args, err := p.psql.
		Update(tableGenres).
		Set(columnName, genre.Name).
		Where(squirrel.Eq{columnID: id}).
		ToSql()
	if err != nil {
		return err
	}

	ct, err := p.pool.Exec(ctx, sql, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return model.ErrGenreAlreadyExists
			}
		}
		return err
	}

	if ct.RowsAffected() == 0 {
		return model.ErrGenreNotFound
	}

	return nil
}

func (p *GenresStorage) RemoveGenre(ctx context.Context, id uuid.UUID) error {
	sql, args, err := p.psql.Delete(tableGenres).Where(squirrel.Eq{columnID: id}).ToSql()
	if err != nil {
		return err
	}
	if _, err = p.pool.Exec(ctx, sql, args...); err != nil {
		return err
	}
	return nil
}

func (p *GenresStorage) ListGenres(ctx context.Context) ([]model.Genre, error) {
	sql, args, err := p.psql.Select(columnID, columnName).From(tableGenres).ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := p.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var genres []model.Genre
	for rows.Next() {
		var genre model.Genre
		if err = rows.Scan(&genre.ID, &genre.Name); err != nil {
			return nil, err
		}
		genres = append(genres, genre)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return genres, nil
}
