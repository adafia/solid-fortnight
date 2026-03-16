CREATE TABLE IF NOT EXISTS evaluation_events (
    id BIGSERIAL PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    flag_key VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    variation_key VARCHAR(255),
    value JSONB,
    reason VARCHAR(255),
    context JSONB,
    evaluated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_evaluation_events_project_env_flag ON evaluation_events(project_id, environment_id, flag_key);
CREATE INDEX IF NOT EXISTS idx_evaluation_events_evaluated_at ON evaluation_events(evaluated_at);
