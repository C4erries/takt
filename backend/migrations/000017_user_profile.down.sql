-- 000017_user_profile.down.sql

ALTER TABLE users DROP COLUMN IF EXISTS avatar_url;
