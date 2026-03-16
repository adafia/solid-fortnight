ALTER TABLE flag_environments DROP CONSTRAINT fk_default_variation;
DROP TABLE IF EXISTS flag_variations;
DROP TABLE IF EXISTS flag_environments;
