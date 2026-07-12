CREATE TABLE IF NOT EXISTS questions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    question_text TEXT NOT NULL,
    normalized_text TEXT NOT NULL,
    question_hash CHAR(64) NOT NULL,
    options_hash CHAR(64) NOT NULL,
    answer_hash CHAR(64) NOT NULL,
    composite_hash CHAR(64) NOT NULL,
    question_type VARCHAR(64) NOT NULL DEFAULT 'other',
    platform VARCHAR(128) NOT NULL DEFAULT '',
    subject VARCHAR(128) NOT NULL DEFAULT '',
    source VARCHAR(255) NOT NULL DEFAULT '',
    collected_at DATETIME(6) NULL,
    status TINYINT NOT NULL DEFAULT 1,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    UNIQUE KEY uk_questions_composite_hash (composite_hash),
    KEY idx_questions_question_hash (question_hash),
    KEY idx_questions_type_subject (question_type, subject),
    FULLTEXT KEY ft_questions_normalized_text (normalized_text)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS question_options (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    question_id BIGINT UNSIGNED NOT NULL,
    option_key VARCHAR(16) NOT NULL,
    option_text TEXT NOT NULL,
    position INT NOT NULL DEFAULT 0,
    UNIQUE KEY uk_question_options_key (question_id, option_key),
    KEY idx_question_options_question (question_id),
    CONSTRAINT fk_question_options_question FOREIGN KEY (question_id) REFERENCES questions (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS question_answers (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    question_id BIGINT UNSIGNED NOT NULL,
    answer_text TEXT NOT NULL,
    answer_raw TEXT NOT NULL,
    position INT NOT NULL DEFAULT 0,
    KEY idx_question_answers_question (question_id),
    CONSTRAINT fk_question_answers_question FOREIGN KEY (question_id) REFERENCES questions (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

