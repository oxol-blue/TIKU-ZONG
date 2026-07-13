CREATE TABLE IF NOT EXISTS answer_feedbacks (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    request_id VARCHAR(64) NOT NULL,
    question_hash CHAR(64) NOT NULL,
    question_text TEXT NOT NULL,
    feedback_type VARCHAR(32) NOT NULL,
    comment VARCHAR(1000) NOT NULL DEFAULT '',
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    UNIQUE KEY uk_answer_feedback_user_request (user_id, request_id),
    KEY idx_answer_feedback_type (feedback_type, created_at),
    CONSTRAINT fk_answer_feedback_user FOREIGN KEY (user_id) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
