DROP INDEX IF EXISTS idx_email_passes_user_id;
DROP INDEX IF EXISTS uq_email_passes_user_id;
DROP INDEX IF EXISTS idx_granted_roles_user_id;

DROP TABLE IF EXISTS email_passes;
DROP TABLE IF EXISTS granted_roles;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS users;