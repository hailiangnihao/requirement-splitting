-- 测试执行记录表
CREATE TABLE test_runs (
    id VARCHAR(100) PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    test_case_id VARCHAR(100) NOT NULL REFERENCES test_cases(id) ON DELETE CASCADE,
    executed_by VARCHAR(100) NOT NULL, -- 可以是 UUID 或者是 "AI"
    execution_type VARCHAR(50) NOT NULL,
    actual_result TEXT,
    evidence JSONB,
    is_defect_suggested BOOLEAN NOT NULL DEFAULT FALSE,
    review_status VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);