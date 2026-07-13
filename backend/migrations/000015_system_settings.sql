CREATE TABLE IF NOT EXISTS system_settings (
    setting_key VARCHAR(64) NOT NULL PRIMARY KEY,
    setting_value TEXT NOT NULL,
    is_public TINYINT NOT NULL DEFAULT 0,
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT IGNORE INTO system_settings (setting_key, setting_value, is_public) VALUES
    ('site_name', '题库调用系统', 1),
    ('support_url', '', 1),
    ('maintenance_notice', '', 1),
    ('registration_enabled', 'true', 1);
