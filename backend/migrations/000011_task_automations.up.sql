-- 000011_task_automations.up.sql

ALTER TABLE automations ADD COLUMN task_id UUID NULL REFERENCES tasks(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_automations_task ON automations(task_id);
