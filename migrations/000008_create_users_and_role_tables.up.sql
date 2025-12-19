CREATE TABLE IF NOT EXISTS users (
  user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS roles (
  role_id SMALLINT PRIMARY KEY,
  name VARCHAR(16) NOT NULL UNIQUE
);

INSERT INTO roles(role_id, name)
VALUES (0, 'user'), (1, 'admin')
ON CONFLICT (role_id) DO NOTHING;

CREATE TABLE IF NOT EXISTS granted_roles (
  user_id UUID REFERENCES users(user_id) ON DELETE CASCADE,
  role_id SMALLINT REFERENCES roles(role_id) ON DELETE CASCADE,
  PRIMARY KEY (user_id, role_id)
);

CREATE INDEX IF NOT EXISTS idx_granted_roles_user_id ON granted_roles(user_id);

CREATE TABLE IF NOT EXISTS email_passes (
  user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  email VARCHAR(254) PRIMARY KEY,
  pass_hash BYTEA NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_email_passes_user_id ON email_passes(user_id);
CREATE INDEX IF NOT EXISTS idx_email_passes_user_id ON email_passes(user_id);