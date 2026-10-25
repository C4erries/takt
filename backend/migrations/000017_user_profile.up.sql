-- 000017_user_profile.up.sql

ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url TEXT NULL;
