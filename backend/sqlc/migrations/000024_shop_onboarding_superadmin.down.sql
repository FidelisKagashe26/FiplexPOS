DROP INDEX IF EXISTS idx_shops_email_unique;
ALTER TABLE shops DROP COLUMN IF EXISTS phone;
ALTER TABLE shops DROP COLUMN IF EXISTS email;
-- PostgreSQL enum values cannot safely be removed from a live enum.
