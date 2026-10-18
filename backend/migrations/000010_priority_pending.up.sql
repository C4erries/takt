-- 000010_priority_pending.up.sql

ALTER TABLE tasks DROP CONSTRAINT IF EXISTS tasks_priority_check;

UPDATE tasks SET priority = 'pending' WHERE priority = 'urgent';

ALTER TABLE tasks ADD CONSTRAINT tasks_priority_check CHECK (priority IN ('none', 'low', 'medium', 'high', 'pending', 'urgent'));

UPDATE automations 
SET rule_definition = REPLACE(rule_definition::text, '"urgent"', '"pending"')::jsonb 
WHERE rule_definition::text LIKE '%"urgent"%';

UPDATE automations 
SET conditions_json = REPLACE(conditions_json::text, '"urgent"', '"pending"')::jsonb 
WHERE conditions_json::text LIKE '%"urgent"%';

UPDATE automations 
SET actions_json = REPLACE(actions_json::text, '"urgent"', '"pending"')::jsonb 
WHERE actions_json::text LIKE '%"urgent"%';
