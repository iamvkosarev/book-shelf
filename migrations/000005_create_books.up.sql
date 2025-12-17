CREATE TABLE IF NOT EXISTS books (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	publisher_id UUID REFERENCES publishers(id) ON DELETE SET NULL,
	published_at DATE,
	title VARCHAR(100) NOT NULL,
	description VARCHAR(1000),
	price DOUBLE PRECISION CHECK (price >= 0)
);

CREATE TABLE IF NOT EXISTS books_authors (
	book_id UUID NOT NULL REFERENCES books(id) ON DELETE CASCADE,
	author_id UUID NOT NULL REFERENCES authors(id) ON DELETE RESTRICT,
	PRIMARY KEY (book_id, author_id)
);

CREATE INDEX IF NOT EXISTS ix_books_authors_author ON books_authors(author_id);

CREATE TABLE IF NOT EXISTS books_tags (
    book_id UUID NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    tag_id  UUID NOT NULL REFERENCES tags(id)  ON DELETE RESTRICT,
    PRIMARY KEY (book_id, tag_id)
);
CREATE INDEX IF NOT EXISTS ix_books_tags_tag ON books_tags(tag_id);