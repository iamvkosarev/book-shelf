CREATE TABLE IF NOT EXISTS authors (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	person_id UUID REFERENCES persons(id),
	pseudonym VARCHAR(100)
);