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

const tableTags = "tags"

type TagsStorage struct {
	pool *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewTagsStorage(pool *pgxpool.Pool) *TagsStorage {
	return &TagsStorage{
		pool: pool,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (p *TagsStorage) AddTag(ctx context.Context, name string) (uuid.UUID, error) {
	sql, args, err := p.psql.Insert(tableTags).Columns(columnName).Values(name).Suffix(
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
				return uuid.Nil, model.ErrTagAlreadyExists
			}
		}
		return uuid.Nil, err
	}
	return id, nil
}

func (p *TagsStorage) GetTag(ctx context.Context, id uuid.UUID) (model.Tag, error) {
	sql, args, err := p.psql.Select(
		columnID, columnName,
	).From(tableTags).Where(squirrel.Eq{columnID: id}).ToSql()
	if err != nil {
		return model.Tag{}, err
	}
	var tag model.Tag
	if err = p.pool.QueryRow(ctx, sql, args...).Scan(&tag.ID, &tag.Name); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Tag{}, model.ErrTagNotFound
		}
		return model.Tag{}, err
	}
	return tag, nil
}

func (p *TagsStorage) UpdateTag(ctx context.Context, id uuid.UUID, tag model.Tag) error {
	sql, args, err := p.psql.
		Update(tableTags).
		Set(columnName, tag.Name).
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
				return model.ErrTagAlreadyExists
			}
		}
		return err
	}

	if ct.RowsAffected() == 0 {
		return model.ErrTagNotFound
	}

	return nil
}

func (p *TagsStorage) RemoveTag(ctx context.Context, id uuid.UUID) error {
	sql, args, err := p.psql.Delete(tableTags).Where(squirrel.Eq{columnID: id}).ToSql()
	if err != nil {
		return err
	}
	if _, err = p.pool.Exec(ctx, sql, args...); err != nil {
		return err
	}
	return nil
}

func (p *TagsStorage) ListTags(ctx context.Context) ([]model.Tag, error) {
	sql, args, err := p.psql.Select(columnID, columnName).From(tableTags).ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := p.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tags []model.Tag
	for rows.Next() {
		var tag model.Tag
		if err = rows.Scan(&tag.ID, &tag.Name); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return tags, nil
}
