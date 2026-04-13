-- +goose Up
SELECT 'up SQL query';
ALTER TABLE accounts DROP CONSTRAINT IF EXISTS accounts_user_id_fkey;
ALTER TABLE accounts DROP COLUMN IF EXISTS user_id;

ALTER TABLE users ADD COLUMN account_id UUID REFERENCES accounts(id);

-- +goose Down
SELECT 'down SQL query';
ALTER TABLE users DROP COLUMN account_id;