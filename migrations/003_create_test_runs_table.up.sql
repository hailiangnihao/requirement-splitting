CREATE TABLE test_runs (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    test_case_id UUID NOT NULL REFERENCES test_cases(id) ON DELETE CASCADE,
    executed_by VARCHAR(255) NOT NULL,
    execution_type VARCHAR(50) NOT NULL,
    actual_result TEXT,
    evidence JSONB,
    is_defect_suggested BOOLEAN DEFAULT FALSE,
    review_status VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
