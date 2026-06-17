DROP INDEX IF EXISTS idx_users_auth_provider;
DROP INDEX IF EXISTS uniq_users_google_sub;

ALTER TABLE users DROP COLUMN IF EXISTS avatar_url;
ALTER TABLE users DROP COLUMN IF EXISTS google_sub;
ALTER TABLE users DROP COLUMN IF EXISTS auth_provider;
