ALTER TABLE flag_environments
ADD COLUMN rollout_percentage DECIMAL(5, 2),
ADD COLUMN rollout_variation_id UUID REFERENCES flag_variations(id) ON DELETE SET NULL;

CREATE TABLE IF NOT EXISTS flag_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    flag_environment_id UUID NOT NULL REFERENCES flag_environments(id) ON DELETE CASCADE,
    variation_id UUID NOT NULL REFERENCES flag_variations(id) ON DELETE CASCADE,
    clauses JSONB NOT NULL, -- Array of clauses: [{attribute, operator, values}]
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(flag_environment_id, id)
);
