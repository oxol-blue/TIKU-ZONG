CREATE TABLE IF NOT EXISTS packages (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(128) NOT NULL,
    package_type VARCHAR(16) NOT NULL,
    duration_seconds BIGINT NULL,
    total_count INT NOT NULL DEFAULT 0,
    ai_count INT NOT NULL DEFAULT 0,
    price_cents INT NOT NULL DEFAULT 0,
    status TINYINT NOT NULL DEFAULT 1,
    limit_count INT NOT NULL DEFAULT 0,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    KEY idx_packages_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS package_instances (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    package_id BIGINT UNSIGNED NOT NULL,
    starts_at DATETIME(6) NOT NULL,
    expires_at DATETIME(6) NULL,
    remaining_count INT NOT NULL DEFAULT 0,
    remaining_ai_count INT NOT NULL DEFAULT 0,
    status TINYINT NOT NULL DEFAULT 1,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    KEY idx_package_instances_user (user_id, status),
    KEY idx_package_instances_expiry (expires_at),
    CONSTRAINT fk_package_instances_user FOREIGN KEY (user_id) REFERENCES users (id),
    CONSTRAINT fk_package_instances_package FOREIGN KEY (package_id) REFERENCES packages (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS package_consumptions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    instance_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    kind VARCHAR(16) NOT NULL,
    amount INT NOT NULL,
    request_id VARCHAR(64) NOT NULL,
    endpoint VARCHAR(128) NOT NULL DEFAULT '',
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    UNIQUE KEY uk_package_consumptions_request (request_id, kind),
    KEY idx_package_consumptions_user (user_id, created_at),
    CONSTRAINT fk_package_consumptions_instance FOREIGN KEY (instance_id) REFERENCES package_instances (id),
    CONSTRAINT fk_package_consumptions_user FOREIGN KEY (user_id) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS api_call_logs (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    request_id VARCHAR(64) NOT NULL,
    user_id BIGINT UNSIGNED NULL,
    api_key_id BIGINT UNSIGNED NULL,
    endpoint VARCHAR(128) NOT NULL,
    question_hash CHAR(64) NOT NULL,
    success TINYINT NOT NULL DEFAULT 0,
    is_ai TINYINT NOT NULL DEFAULT 0,
    elapsed_micros BIGINT NOT NULL DEFAULT 0,
    http_status SMALLINT NOT NULL,
    error_code VARCHAR(64) NOT NULL DEFAULT '',
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    UNIQUE KEY uk_api_call_logs_request (request_id),
    KEY idx_api_call_logs_user_time (user_id, created_at),
    CONSTRAINT fk_api_call_logs_user FOREIGN KEY (user_id) REFERENCES users (id),
    CONSTRAINT fk_api_call_logs_key FOREIGN KEY (api_key_id) REFERENCES user_api_keys (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

