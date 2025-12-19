BEGIN;

CREATE TABLE IF NOT EXISTS persons (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	first_name VARCHAR(100) NOT NULL,
	last_name VARCHAR(100) NOT NULL,
	middle_name VARCHAR(100)
);

ALTER TABLE authors ADD COLUMN IF NOT EXISTS person_id UUID;

INSERT INTO persons (id, first_name, last_name, middle_name)
SELECT
	a.id,
	COALESCE(a.first_name, a.pseudonym, 'null') AS first_name,
	COALESCE(a.last_name, 'null') AS last_name,
	a.middle_name
FROM authors a ON CONFLICT (id) DO NOTHING;

UPDATE authors SET person_id = id WHERE person_id IS NULL;

ALTER TABLE authors ADD CONSTRAINT authors_person_id_fkey
	FOREIGN KEY (person_id) REFERENCES persons(id);

ALTER TABLE authors DROP CONSTRAINT IF EXISTS chk_authors_identity;

ALTER TABLE authors
	DROP COLUMN IF EXISTS first_name,
	DROP COLUMN IF EXISTS last_name,
	DROP COLUMN IF EXISTS middle_name;
COMMIT;