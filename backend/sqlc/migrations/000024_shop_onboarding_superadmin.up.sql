-- A shop is the tenant and owns its public contact details. An account only
-- authenticates a person working in that shop.
ALTER TYPE user_role ADD VALUE IF NOT EXISTS 'superadmin';

ALTER TABLE shops
  ADD COLUMN IF NOT EXISTS email varchar,
  ADD COLUMN IF NOT EXISTS phone varchar;

CREATE UNIQUE INDEX IF NOT EXISTS idx_shops_email_unique
  ON shops (lower(email)) WHERE email IS NOT NULL;
