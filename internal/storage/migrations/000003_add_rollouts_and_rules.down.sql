DROP TABLE IF EXISTS flag_rules;

ALTER TABLE flag_environments
DROP COLUMN IF EXISTS rollout_percentage,
DROP COLUMN IF EXISTS rollout_variation_id;
