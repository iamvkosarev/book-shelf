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

const (
	tablePersons = "persons"

	columnFirstName  = "first_name"
	columnLastName   = "last_name"
	columnMiddleName = "middle_name"
)

type PersonsStorage struct {
	pool *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewPersonsStorage(pool *pgxpool.Pool) *PersonsStorage {
	return &PersonsStorage{
		pool: pool,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (p *PersonsStorage) AddPerson(ctx context.Context, firstName, lastName, middleName string) (uuid.UUID, error) {
	sql, args, err := p.psql.Insert(tablePersons).Columns(
		columnFirstName, columnLastName, columnMiddleName,
	).Values(firstName, lastName, middleName).Suffix(
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
				return uuid.Nil, model.ErrPersonAlreadyExists
			}
		}
		return uuid.Nil, err
	}
	return id, nil
}

func (p *PersonsStorage) GetPerson(ctx context.Context, id uuid.UUID) (model.Person, error) {
	sql, args, err := p.psql.Select(
		columnID, columnFirstName, columnLastName, columnMiddleName,
	).From(tablePersons).Where(squirrel.Eq{columnID: id}).ToSql()
	if err != nil {
		return model.Person{}, err
	}
	var person model.Person
	if err = p.pool.QueryRow(ctx, sql, args...).Scan(
		&person.ID, &person.FirstName, &person.LastName,
		&person.MiddleName,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Person{}, model.ErrPersonNotFound
		}
		return model.Person{}, err
	}
	return person, nil
}

func (p *PersonsStorage) UpdatePerson(ctx context.Context, id uuid.UUID, person model.Person) error {
	sql, args, err := p.psql.
		Update(tablePersons).
		Set(columnFirstName, person.FirstName).
		Set(columnLastName, person.LastName).
		Set(columnMiddleName, person.MiddleName).
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
				return model.ErrPersonAlreadyExists
			}
		}
		return err
	}

	if ct.RowsAffected() == 0 {
		return model.ErrPersonNotFound
	}

	return nil
}

func (p *PersonsStorage) RemovePerson(ctx context.Context, id uuid.UUID) error {
	sql, args, err := p.psql.Delete(tablePersons).Where(squirrel.Eq{columnID: id}).ToSql()
	if err != nil {
		return err
	}
	if _, err = p.pool.Exec(ctx, sql, args...); err != nil {
		return err
	}
	return nil
}

func (p *PersonsStorage) ListPersons(ctx context.Context) ([]model.Person, error) {
	sql, args, err := p.psql.Select(
		columnID, columnFirstName, columnLastName, columnMiddleName,
	).From(tablePersons).ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := p.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var persons []model.Person
	for rows.Next() {
		var person model.Person
		if err = rows.Scan(&person.ID, &person.FirstName, &person.LastName, &person.MiddleName); err != nil {
			return nil, err
		}
		persons = append(persons, person)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return persons, nil
}
