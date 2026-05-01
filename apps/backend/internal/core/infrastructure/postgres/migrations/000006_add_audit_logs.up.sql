-- Create the parent table, partitioned by the creation date
CREATE TABLE audit_logs (
    id UUID NOT NULL,
    group_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    action VARCHAR(255) NOT NULL,
    target_id VARCHAR(255),
    details TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id, created_at) -- Partition key MUST be in the primary key
) PARTITION BY RANGE (created_at);

-- create initial partition for the current month so the app can start immediately
-- use a dynamic naming convention: audit_logs_yYYYYmMM
CREATE TABLE audit_logs_y2026m04 PARTITION OF audit_logs 
    FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');

-- index automatically cascades to all underlying partitions
CREATE INDEX idx_audit_logs_group_id ON audit_logs(group_id, created_at DESC);