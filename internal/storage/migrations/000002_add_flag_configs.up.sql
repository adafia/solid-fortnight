CREATE TABLE IF NOT EXISTS flag_environments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    flag_id UUID NOT NULL REFERENCES flags(id) ON DELETE CASCADE,
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    enabled BOOLEAN DEFAULT FALSE,
    default_variation_id UUID, -- References flag_variations.id later
    version INT DEFAULT 1,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID,
    UNIQUE(flag_id, environment_id)
);

CREATE TABLE IF NOT EXISTS flag_variations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    flag_environment_id UUID NOT NULL REFERENCES flag_environments(id) ON DELETE CASCADE,
    key VARCHAR(255) NOT NULL,
    value JSONB NOT NULL,
    name VARCHAR(255),
    description TEXT,
    UNIQUE(flag_environment_id, key)
);

-- Add foreign key from flag_environments to flag_variations once flag_variations exists
ALTER TABLE flag_environments
ADD CONSTRAINT fk_default_variation
FOREIGN KEY (default_variation_id)
REFERENCES flag_variations(id)
ON DELETE SET NULL;
