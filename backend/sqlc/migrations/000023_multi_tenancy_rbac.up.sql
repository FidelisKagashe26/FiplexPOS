-- Multi-tenancy + RBAC (additive, non-destructive).
--
-- NOTE: we intentionally KEEP users.role (the legacy enum) as a coarse
-- "account tier". The new roles/permissions tables provide fine-grained,
-- per-shop authorization on top of it (mirrors SMAS's account_type + role.permissions
-- two-tier model). users with role = 'admin' act as the privileged owner tier and
-- bypass fine-grained permission checks. This keeps existing auth working and makes
-- this migration safe to run on the live database (AUTO_MIGRATE=true).

CREATE TABLE shops (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name varchar NOT NULL,
  address varchar,
  owner_id uuid NOT NULL REFERENCES users(id),
  is_active boolean NOT NULL DEFAULT true,
  created_at timestamp with time zone NOT NULL DEFAULT now(),
  updated_at timestamp with time zone NOT NULL DEFAULT now()
);

CREATE TABLE permissions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name varchar NOT NULL UNIQUE,
  description text,
  module varchar NOT NULL,
  created_at timestamp with time zone NOT NULL DEFAULT now()
);

CREATE TABLE roles (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  shop_id uuid REFERENCES shops(id) ON DELETE CASCADE,
  name varchar NOT NULL,
  description text,
  is_active boolean NOT NULL DEFAULT true,
  created_at timestamp with time zone NOT NULL DEFAULT now(),
  updated_at timestamp with time zone NOT NULL DEFAULT now(),
  UNIQUE (shop_id, name)
);

CREATE TABLE role_permissions (
  role_id uuid NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  permission_id uuid NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
  PRIMARY KEY (role_id, permission_id)
);

-- user_shop_roles: a user's role within a shop. Named to avoid the Go identifier
-- collision between the user_role enum (kept) and a "user_roles" table struct.
CREATE TABLE user_shop_roles (
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  shop_id uuid REFERENCES shops(id) ON DELETE CASCADE,
  role_id uuid NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  PRIMARY KEY (user_id, shop_id)
);

-- Indexes to keep the per-request permission resolver fast.
CREATE INDEX idx_user_shop_roles_user_id ON user_shop_roles (user_id);
CREATE INDEX idx_roles_shop_id ON roles (shop_id);
CREATE INDEX idx_shops_owner_id ON shops (owner_id);

-- Add shop_id to existing tables (nullable so existing single-tenant rows stay valid).
ALTER TABLE products ADD COLUMN shop_id uuid REFERENCES shops(id) ON DELETE CASCADE;
ALTER TABLE orders ADD COLUMN shop_id uuid REFERENCES shops(id) ON DELETE CASCADE;
ALTER TABLE categories ADD COLUMN shop_id uuid REFERENCES shops(id) ON DELETE CASCADE;
ALTER TABLE promotions ADD COLUMN shop_id uuid REFERENCES shops(id) ON DELETE CASCADE;
ALTER TABLE shifts ADD COLUMN shop_id uuid REFERENCES shops(id) ON DELETE CASCADE;
ALTER TABLE cash_transactions ADD COLUMN shop_id uuid REFERENCES shops(id) ON DELETE CASCADE;
ALTER TABLE customers ADD COLUMN shop_id uuid REFERENCES shops(id) ON DELETE CASCADE;
ALTER TABLE activity_logs ADD COLUMN shop_id uuid REFERENCES shops(id) ON DELETE CASCADE;
