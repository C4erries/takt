-- 000009_low_code_automations_dsl.up.sql

ALTER TABLE automations ADD COLUMN rule_definition JSONB NOT NULL DEFAULT '{}'::jsonb;
