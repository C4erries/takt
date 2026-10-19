-- 000011_task_automations.down.sql

DROP INDEX IF EXISTS idx_automations_task;
ALTER TABLE automations DROP COLUMN IF EXISTS task_id;
