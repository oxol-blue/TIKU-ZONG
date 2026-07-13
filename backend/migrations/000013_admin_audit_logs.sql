CREATE TABLE IF NOT EXISTS admin_audit_logs (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    admin_id BIGINT UNSIGNED NOT NULL,
    admin_email VARCHAR(255) NOT NULL,
    action VARCHAR(32) NOT NULL,
    resource VARCHAR(255) NOT NULL,
    request_path VARCHAR(512) NOT NULL,
    ip_address VARCHAR(64) NOT NULL DEFAULT '',
    http_status SMALLINT UNSIGNED NOT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    KEY idx_admin_audit_logs_created (created_at),
    KEY idx_admin_audit_logs_admin_created (admin_id, created_at),
    KEY idx_admin_audit_logs_resource_created (resource, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
