ALTER TABLE users ADD COLUMN IF NOT EXISTS auth_provider VARCHAR NOT NULL DEFAULT 'password';
ALTER TABLE users ADD COLUMN IF NOT EXISTS google_sub VARCHAR;
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS uniq_users_google_sub
ON users(google_sub)
WHERE google_sub IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_users_auth_provider
ON users(auth_provider);
