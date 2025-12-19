ALTER TABLE books_tags
  DROP CONSTRAINT IF EXISTS books_tags_tag_id_fkey;

ALTER TABLE books_tags
  ADD CONSTRAINT books_tags_tag_id_fkey
  FOREIGN KEY (tag_id)
  REFERENCES tags(id)
  ON DELETE RESTRICT;