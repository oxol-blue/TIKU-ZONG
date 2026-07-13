ALTER TABLE api_call_logs ADD COLUMN source_kind VARCHAR(16) NOT NULL DEFAULT '' AFTER is_ai;
ALTER TABLE api_call_logs ADD KEY idx_api_call_logs_source_success (source_kind, success, created_at);
