package postgres

import (
	"context"
	"errors"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/iamvkosarev/book-shelf/internal/model"
	"github.com/iamvkosarev/book-shelf/internal/usecase/books"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
)

const (
	tableBooks        = "books"
	tableBooksAuthors = "books_authors"
	tableBooksTags    = "books_tags"

	columnPublisherID = "publisher_id"
	columnPublishedAt = "published_at"
	columnTitle       = "title"
	columnDescription = "description"
	columnPrice       = "price"
	columnMark        = "mark"

	columnBookID   = "book_id"
	columnAuthorID = "author_id"
	columnTagID    = "tag_id"
)

type BooksStorage struct {
	pool *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewBooksStorage(pool *pgxpool.Pool) *BooksStorage {
	return &BooksStorage{
		pool: pool,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (p *BooksStorage) AddBook(ctx context.Context, input books.CreateBookInput) (uuid.UUID, error) {
	if strings.TrimSpace(input.Title) == "" {
		return uuid.Nil, model.ErrBookInvalidFields
	}

	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	sql, args, err := p.psql.Insert(tableBooks).
		Columns(
			columnPublisherID,
			columnPublishedAt,
			columnTitle,
			columnDescription,
			columnPrice,
			columnMark,
		).
		Values(
			toPostgresUUIDPtr(input.PublisherID),
			toPostgresDatePtr(input.PublishedAt),
			input.Title,
			toPostgresTextPtr(input.Description),
			toPostgresFloat8Ptr(input.Price),
			toPostgresInt2Ptr(input.Mark),
		).
		Suffix("RETURNING " + columnID).
		ToSql()
	if err != nil {
		return uuid.Nil, err
	}

	var id uuid.UUID
	if err = tx.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
		return uuid.Nil, err
	}

	if err = p.replaceBookAuthorsTx(ctx, tx, id, input.AuthorsIDs, true); err != nil {
		return uuid.Nil, err
	}
	if err = p.replaceBookTagsTx(ctx, tx, id, input.TagsIDs, true); err != nil {
		return uuid.Nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (p *BooksStorage) GetBook(ctx context.Context, id uuid.UUID) (model.Book, error) {
	sql, args, err := p.psql.Select(
		columnID,
		columnPublisherID,
		columnPublishedAt,
		columnTitle,
		columnDescription,
		columnPrice,
		columnMark,
	).From(tableBooks).Where(squirrel.Eq{columnID: id}).ToSql()
	if err != nil {
		return model.Book{}, err
	}

	var (
		book        model.Book
		publisherID pgtype.UUID
		publishedAt pgtype.Date
		description pgtype.Text
		price       pgtype.Float8
		mark        pgtype.Int2
	)

	if err = p.pool.QueryRow(ctx, sql, args...).Scan(
		&book.ID,
		&publisherID,
		&publishedAt,
		&book.Title,
		&description,
		&price,
		&mark,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Book{}, model.ErrBookNotFound
		}
		return model.Book{}, err
	}

	if publisherID.Valid {
		v := uuid.UUID(publisherID.Bytes)
		book.PublisherID = &v
	} else {
		book.PublisherID = nil
	}

	if publishedAt.Valid {
		t := publishedAt.Time
		book.PublishedAt = &t
	} else {
		book.PublishedAt = nil
	}

	if description.Valid {
		s := description.String
		book.Description = &s
	} else {
		book.Description = nil
	}

	if price.Valid {
		f := price.Float64
		book.Price = &f
	} else {
		book.Price = nil
	}

	if mark.Valid {
		f := mark.Int16
		book.Mark = &f
	} else {
		book.Mark = nil
	}

	authorsIDs, err := p.getBookAuthorIDs(ctx, id)
	if err != nil {
		return model.Book{}, err
	}
	book.AuthorsIDs = authorsIDs

	tagsIDs, err := p.getBookTagIDs(ctx, id)
	if err != nil {
		return model.Book{}, err
	}
	book.TagsIDs = tagsIDs

	return book, nil
}

func (p *BooksStorage) UpdateBook(ctx context.Context, id uuid.UUID, patch books.UpdateBookPatch) error {
	if patch.PublisherID == nil &&
		patch.PublishedAt == nil &&
		patch.Title == nil &&
		patch.Description == nil &&
		patch.Price == nil &&
		patch.Mark == nil &&
		patch.AuthorsIDs == nil &&
		patch.TagsIDs == nil {
		return model.ErrBookInvalidFields
	}

	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	upd := p.psql.Update(tableBooks).Where(squirrel.Eq{columnID: id})

	changed := false

	if patch.PublisherID != nil {
		upd = upd.Set(columnPublisherID, toPostgresUUIDPtr(patch.PublisherID))
		changed = true
	}
	if patch.PublishedAt != nil {
		upd = upd.Set(columnPublishedAt, toPostgresDatePtr(patch.PublishedAt))
		changed = true
	}
	if patch.Title != nil {
		if strings.TrimSpace(*patch.Title) == "" {
			return model.ErrBookInvalidFields
		}
		upd = upd.Set(columnTitle, *patch.Title)
		changed = true
	}
	if patch.Description != nil {
		upd = upd.Set(columnDescription, toPostgresTextPtr(patch.Description))
		changed = true
	}
	if patch.Price != nil {
		upd = upd.Set(columnPrice, toPostgresFloat8Ptr(patch.Price))
		changed = true
	}
	if patch.Mark != nil {
		upd = upd.Set(columnMark, toPostgresInt2Ptr(patch.Mark))
		changed = true
	}

	if changed {
		sql, args, err := upd.ToSql()
		if err != nil {
			return err
		}

		commandTag, err := tx.Exec(ctx, sql, args...)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return model.ErrBookAlreadyExists
			}
			return err
		}
		if commandTag.RowsAffected() == 0 {
			return model.ErrBookNotFound
		}
	} else {
		if _, err := p.ensureBookExistsTx(ctx, tx, id); err != nil {
			return err
		}
	}

	if patch.AuthorsIDs != nil {
		if err := p.replaceBookAuthorsTx(ctx, tx, id, patch.AuthorsIDs, true); err != nil {
			return err
		}
	}
	if patch.TagsIDs != nil {
		if err := p.replaceBookTagsTx(ctx, tx, id, patch.TagsIDs, true); err != nil {
			return err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (p *BooksStorage) RemoveBook(ctx context.Context, id uuid.UUID) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, "DELETE FROM "+tableBooksAuthors+" WHERE "+columnBookID+"=$1", id); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "DELETE FROM "+tableBooksTags+" WHERE "+columnBookID+"=$1", id); err != nil {
		return err
	}

	commandTag, err := tx.Exec(ctx, "DELETE FROM "+tableBooks+" WHERE "+columnID+"=$1", id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return model.ErrBookNotFound
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (p *BooksStorage) ListBooks(ctx context.Context, parameters books.ListBookParameters) ([]model.Book, error) {
	q := p.psql.Select(
		columnID,
		columnPublisherID,
		columnPublishedAt,
		columnTitle,
		columnDescription,
		columnPrice,
		columnMark,
	).From(tableBooks)

	if parameters.AuthorsIDs != nil && len(parameters.AuthorsIDs) > 0 {
		sub := p.psql.Select(columnBookID).
			From(tableBooksAuthors).
			Where(squirrel.Eq{columnAuthorID: parameters.AuthorsIDs}).
			GroupBy(columnBookID).
			Having("COUNT(DISTINCT "+columnAuthorID+") = ?", len(parameters.AuthorsIDs))

		q = q.Where(squirrel.Expr(columnID+" IN (?)", sub))
	}

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := p.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []model.Book
	var ids []uuid.UUID

	for rows.Next() {
		var (
			book        model.Book
			publisherID pgtype.UUID
			publishedAt pgtype.Date
			description pgtype.Text
			price       pgtype.Float8
			mark        pgtype.Int2
		)

		if err := rows.Scan(
			&book.ID,
			&publisherID,
			&publishedAt,
			&book.Title,
			&description,
			&price,
			&mark,
		); err != nil {
			return nil, err
		}

		if publisherID.Valid {
			v := uuid.UUID(publisherID.Bytes)
			book.PublisherID = &v
		}
		if publishedAt.Valid {
			t := publishedAt.Time
			book.PublishedAt = &t
		}
		if description.Valid {
			s := description.String
			book.Description = &s
		}
		if price.Valid {
			f := price.Float64
			book.Price = &f
		}

		if mark.Valid {
			f := mark.Int16
			book.Mark = &f
		}

		books = append(books, book)
		ids = append(ids, book.ID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return books, nil
	}

	authorsMap, err := p.listAuthorsByBookIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	tagsMap, err := p.listTagsByBookIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	for i := range books {
		books[i].AuthorsIDs = authorsMap[books[i].ID]
		books[i].TagsIDs = tagsMap[books[i].ID]
	}

	return books, nil
}

func (p *BooksStorage) ensureBookExistsTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) (uuid.UUID, error) {
	var existing uuid.UUID
	err := tx.QueryRow(ctx, "SELECT "+columnID+" FROM "+tableBooks+" WHERE "+columnID+"=$1", id).Scan(&existing)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, model.ErrBookNotFound
		}
		return uuid.Nil, err
	}
	return existing, nil
}

func (p *BooksStorage) getBookAuthorIDs(ctx context.Context, bookID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := p.pool.Query(
		ctx,
		"SELECT "+columnAuthorID+" FROM "+tableBooksAuthors+" WHERE "+columnBookID+"=$1",
		bookID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (p *BooksStorage) getBookTagIDs(ctx context.Context, bookID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := p.pool.Query(
		ctx,
		"SELECT "+columnTagID+" FROM "+tableBooksTags+" WHERE "+columnBookID+"=$1",
		bookID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (p *BooksStorage) replaceBookAuthorsTx(
	ctx context.Context,
	tx pgx.Tx,
	bookID uuid.UUID,
	authorIDs []uuid.UUID,
	force bool,
) error {
	if force {
		if _, err := tx.Exec(ctx, "DELETE FROM "+tableBooksAuthors+" WHERE "+columnBookID+"=$1", bookID); err != nil {
			return err
		}
	}

	if len(authorIDs) == 0 {
		return nil
	}

	ins := p.psql.Insert(tableBooksAuthors).Columns(columnBookID, columnAuthorID)
	for _, aID := range authorIDs {
		ins = ins.Values(bookID, aID)
	}
	sql, args, err := ins.ToSql()
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, sql, args...); err != nil {
		return err
	}
	return nil
}

func (p *BooksStorage) replaceBookTagsTx(
	ctx context.Context,
	tx pgx.Tx,
	bookID uuid.UUID,
	tagIDs []uuid.UUID,
	force bool,
) error {
	if force {
		if _, err := tx.Exec(ctx, "DELETE FROM "+tableBooksTags+" WHERE "+columnBookID+"=$1", bookID); err != nil {
			return err
		}
	}

	if len(tagIDs) == 0 {
		return nil
	}

	ins := p.psql.Insert(tableBooksTags).Columns(columnBookID, columnTagID)
	for _, tID := range tagIDs {
		ins = ins.Values(bookID, tID)
	}
	sql, args, err := ins.ToSql()
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, sql, args...); err != nil {
		return err
	}
	return nil
}

func (p *BooksStorage) listAuthorsByBookIDs(ctx context.Context, bookIDs []uuid.UUID) (
	map[uuid.UUID][]uuid.UUID,
	error,
) {
	result := make(map[uuid.UUID][]uuid.UUID)
	if len(bookIDs) == 0 {
		return result, nil
	}

	rows, err := p.pool.Query(
		ctx,
		"SELECT "+columnBookID+", "+columnAuthorID+" FROM "+tableBooksAuthors+" WHERE "+columnBookID+" = ANY($1)",
		bookIDs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var bID, aID uuid.UUID
		if err := rows.Scan(&bID, &aID); err != nil {
			return nil, err
		}
		result[bID] = append(result[bID], aID)
	}
	return result, rows.Err()
}

func (p *BooksStorage) listTagsByBookIDs(ctx context.Context, bookIDs []uuid.UUID) (map[uuid.UUID][]uuid.UUID, error) {
	result := make(map[uuid.UUID][]uuid.UUID)
	if len(bookIDs) == 0 {
		return result, nil
	}

	rows, err := p.pool.Query(
		ctx,
		"SELECT "+columnBookID+", "+columnTagID+" FROM "+tableBooksTags+" WHERE "+columnBookID+" = ANY($1)",
		bookIDs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var bID, tID uuid.UUID
		if err := rows.Scan(&bID, &tID); err != nil {
			return nil, err
		}
		result[bID] = append(result[bID], tID)
	}
	return result, rows.Err()
}
