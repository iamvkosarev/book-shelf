package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/iamvkosarev/book-shelf/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
)

const (
	tableEmailPassed = "email_passes"
	tableUsers       = "users"

	RoleAdmin      = "admin"
	RoleUser       = "user"
	columnUserID   = "user_id"
	columnEmail    = "email"
	columnPassHash = "pass_hash"
)

type UserStorage struct {
	pool *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewUsersStorage(pool *pgxpool.Pool) *UserStorage {
	return &UserStorage{
		pool: pool,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (u *UserStorage) GetUserRolesByID(ctx context.Context, userID uuid.UUID) ([]model.Role, error) {
	sql, args, err := u.psql.
		Select("r.name").
		From("granted_roles gr").
		Join("roles r ON r.role_id = gr.role_id").
		Where(squirrel.Eq{"gr.user_id": userID}).
		OrderBy("r.role_id").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	rows, err := u.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("exec query: %w", err)
	}
	defer rows.Close()

	var roles []model.Role
	for rows.Next() {
		var dbName string
		if err = rows.Scan(&dbName); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		switch strings.ToLower(dbName) {
		case RoleAdmin:
			roles = append(roles, model.RoleAdmin)
		case RoleUser:
			roles = append(roles, model.RoleUser)
		default:
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	return roles, nil
}

func (u *UserStorage) GetUserByID(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	query, args, err := u.psql.
		Select("u.user_id", "u.created_at", "ep.email").
		From("users u").
		LeftJoin("email_passes ep ON u.user_id = ep.user_id").
		Where(squirrel.Eq{"u.user_id": userID}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var (
		user   model.User
		emailN sql.NullString
	)
	if err = u.pool.QueryRow(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &emailN); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrUserNotFound
		}
		return nil, fmt.Errorf("exec query: %w", err)
	}
	if emailN.Valid {
		user.Email = emailN.String
	} else {
		user.Email = ""
	}
	return &user, nil
}

func (u *UserStorage) GetIDAndPassHash(ctx context.Context, email string) (uuid.UUID, []byte, error) {
	query, args, err := u.psql.
		Select(columnUserID, columnPassHash).
		From(tableEmailPassed).
		Where(squirrel.Eq{columnEmail: email}).
		ToSql()
	if err != nil {
		return uuid.Nil, nil, err
	}

	var (
		userID   uuid.UUID
		passHash []byte
	)

	if err = u.pool.QueryRow(ctx, query, args...).Scan(&userID, &passHash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, nil, model.ErrUserNotFound
		}
		return uuid.Nil, nil, fmt.Errorf("exec query: %w", err)
	}

	return userID, passHash, nil
}

func (u *UserStorage) CreateAnonymouseUser(ctx context.Context) (uuid.UUID, error) {
	query, args, err := u.psql.
		Insert(tableUsers).
		Columns(columnUserID).
		Values(squirrel.Expr("DEFAULT")).
		Suffix("RETURNING user_id").
		ToSql()
	if err != nil {
		return uuid.Nil, err
	}

	var userID uuid.UUID
	if err = u.pool.QueryRow(ctx, query, args...).Scan(&userID); err != nil {
		return uuid.Nil, fmt.Errorf("exec query: %w", err)
	}

	return userID, nil
}

func (u *UserStorage) CreateUserByEmail(ctx context.Context, email string, hash []byte) (uuid.UUID, error) {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q1, a1, err := u.psql.
		Insert(tableUsers).
		Columns(columnUserID).
		Values(squirrel.Expr("DEFAULT")).
		Suffix("RETURNING user_id").
		ToSql()
	if err != nil {
		return uuid.Nil, err
	}

	var userID uuid.UUID
	if err = tx.QueryRow(ctx, q1, a1...).Scan(&userID); err != nil {
		return uuid.Nil, fmt.Errorf("exec users insert: %w", err)
	}

	q2, a2, err := u.psql.
		Insert(tableEmailPassed).
		Columns(columnUserID, columnEmail, columnPassHash).
		Values(userID, email, hash).
		ToSql()
	if err != nil {
		return uuid.Nil, err
	}

	if _, err = tx.Exec(ctx, q2, a2...); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return uuid.Nil, model.ErrUserAlreadyExists
		}
		return uuid.Nil, fmt.Errorf("exec email_passes insert: %w", err)
	}

	q3, a3, err := u.psql.
		Insert("granted_roles").
		Columns("user_id", "role_id").
		Values(userID, 0).
		ToSql()
	if err != nil {
		return uuid.Nil, err
	}

	if _, err = tx.Exec(ctx, q3, a3...); err != nil {
		return uuid.Nil, fmt.Errorf("exec granted_roles insert: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return uuid.Nil, fmt.Errorf("commit tx: %w", err)
	}
	return userID, nil
}
