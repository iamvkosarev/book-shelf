BEGIN;

ALTER TABLE authors
	ADD COLUMN IF NOT EXISTS first_name VARCHAR(100),
	ADD COLUMN IF NOT EXISTS last_name VARCHAR(100),
	ADD COLUMN IF NOT EXISTS middle_name VARCHAR(100);

UPDATE authors a SET
	first_name = p.first_name,
	last_name = p.last_name,
	middle_name = p.middle_name
FROM persons p WHERE a.person_id = p.id;

ALTER TABLE authors DROP CONSTRAINT IF EXISTS authors_person_id_fkey;

ALTER TABLE authors DROP COLUMN IF EXISTS person_id;

ALTER TABLE authors ADD CONSTRAINT chk_authors_identity
	CHECK (
        pseudonym IS NOT NULL
        OR first_name IS NOT NULL
        OR last_name IS NOT NULL
  );

DROP TABLE IF EXISTS persons;

COMMIT;