-- 000010_priority_pending.down.sql

ALTER TABLE tasks DROP CONSTRAINT IF EXISTS tasks_priority_check;

UPDATE tasks SET priority = 'urgent' WHERE priority = 'pending';

ALTER TABLE tasks ADD CONSTRAINT tasks_priority_check CHECK (priority IN ('none', 'low', 'medium', 'high', 'urgent'));

UPDATE automations 
SET rule_definition = REPLACE(rule_definition::text, '"pending"', '"urgent"')::jsonb 
WHERE rule_definition::text LIKE '%"pending"%';

UPDATE automations 
SET conditions_json = REPLACE(conditions_json::text, '"pending"', '"urgent"')::jsonb 
WHERE conditions_json::text LIKE '%"pending"%';

UPDATE automations 
SET actions_json = REPLACE(actions_json::text, '"pending"', '"urgent"')::jsonb 
WHERE actions_json::text LIKE '%"pending"%';
