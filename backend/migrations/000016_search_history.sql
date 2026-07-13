CREATE TABLE IF NOT EXISTS search_history (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    request_id VARCHAR(96) NOT NULL,
    question_text TEXT NOT NULL,
    question_type VARCHAR(64) NOT NULL DEFAULT '',
    answer_text TEXT NOT NULL,
    source VARCHAR(255) NOT NULL DEFAULT '',
    is_ai TINYINT NOT NULL DEFAULT 0,
    elapsed_micros BIGINT NOT NULL DEFAULT 0,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    UNIQUE KEY uk_search_history_user_request (user_id, request_id),
    KEY idx_search_history_user_created (user_id, created_at),
    KEY idx_search_history_user_ai_created (user_id, is_ai, created_at),
    CONSTRAINT fk_search_history_user FOREIGN KEY (user_id) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
