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

const tablePublishers = "publishers"

type PublishersStorage struct {
	pool *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewPublishersStorage(pool *pgxpool.Pool) *PublishersStorage {
	return &PublishersStorage{
		pool: pool,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (p *PublishersStorage) AddPublisher(ctx context.Context, name string) (uuid.UUID, error) {
	sql, args, err := p.psql.Insert(tablePublishers).Columns(columnName).Values(name).Suffix(
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
				return uuid.Nil, model.ErrPublisherAlreadyExists
			}
		}
		return uuid.Nil, err
	}
	return id, nil
}

func (p *PublishersStorage) GetPublisher(ctx context.Context, id uuid.UUID) (model.Publisher, error) {
	sql, args, err := p.psql.Select(
		columnID, columnName,
	).From(tablePublishers).Where(squirrel.Eq{columnID: id}).ToSql()
	if err != nil {
		return model.Publisher{}, err
	}
	var publisher model.Publisher
	if err = p.pool.QueryRow(ctx, sql, args...).Scan(&publisher.ID, &publisher.Name); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Publisher{}, model.ErrPublisherNotFound
		}
		return model.Publisher{}, err
	}
	return publisher, nil
}

func (p *PublishersStorage) UpdatePublisher(ctx context.Context, id uuid.UUID, publisher model.Publisher) error {
	sql, args, err := p.psql.
		Update(tablePublishers).
		Set(columnName, publisher.Name).
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
				return model.ErrPublisherAlreadyExists
			}
		}
		return err
	}

	if ct.RowsAffected() == 0 {
		return model.ErrPublisherNotFound
	}

	return nil
}

func (p *PublishersStorage) RemovePublisher(ctx context.Context, id uuid.UUID) error {
	sql, args, err := p.psql.Delete(tablePublishers).Where(squirrel.Eq{columnID: id}).ToSql()
	if err != nil {
		return err
	}
	if _, err = p.pool.Exec(ctx, sql, args...); err != nil {
		return err
	}
	return nil
}

func (p *PublishersStorage) ListPublishers(ctx context.Context) ([]model.Publisher, error) {
	sql, args, err := p.psql.Select(columnID, columnName).From(tablePublishers).ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := p.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var publishers []model.Publisher
	for rows.Next() {
		var publisher model.Publisher
		if err = rows.Scan(&publisher.ID, &publisher.Name); err != nil {
			return nil, err
		}
		publishers = append(publishers, publisher)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return publishers, nil
}
