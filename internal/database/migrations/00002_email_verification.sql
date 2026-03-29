-- +goose Up
ALTER TABLE users
  ADD COLUMN email_verified                BOOLEAN     NOT NULL DEFAULT false,
  ADD COLUMN verification_token            TEXT        UNIQUE,
  ADD COLUMN verification_token_expires_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE users
  DROP COLUMN IF EXISTS verification_token_expires_at,
  DROP COLUMN IF EXISTS verification_token,
  DROP COLUMN IF EXISTS email_verified;
