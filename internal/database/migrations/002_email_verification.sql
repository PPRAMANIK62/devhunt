ALTER TABLE users
  ADD COLUMN email_verified                BOOLEAN     NOT NULL DEFAULT false,
  ADD COLUMN verification_token            TEXT        UNIQUE,
  ADD COLUMN verification_token_expires_at TIMESTAMPTZ;
