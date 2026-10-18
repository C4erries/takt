-- 000009_low_code_automations_dsl.down.sql

ALTER TABLE automations DROP COLUMN IF EXISTS rule_definition;
