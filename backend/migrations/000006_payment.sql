CREATE TABLE IF NOT EXISTS payment_gateways (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    provider VARCHAR(32) NOT NULL,
    name VARCHAR(128) NOT NULL,
    base_url VARCHAR(512) NOT NULL,
    merchant_id VARCHAR(128) NOT NULL,
    secret_ciphertext TEXT NOT NULL,
    enabled TINYINT NOT NULL DEFAULT 0,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    UNIQUE KEY uk_payment_gateways_provider (provider)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS payment_orders (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    order_no VARCHAR(64) NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    package_id BIGINT UNSIGNED NOT NULL,
    provider VARCHAR(32) NOT NULL,
    amount_cents INT NOT NULL,
    payable_cents INT NOT NULL,
    refunded_cents INT NOT NULL DEFAULT 0,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    provider_trade_no VARCHAR(128) NOT NULL DEFAULT '',
    package_instance_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
    expires_at DATETIME(6) NOT NULL,
    paid_at DATETIME(6) NULL,
    closed_at DATETIME(6) NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    UNIQUE KEY uk_payment_orders_no (order_no),
    KEY idx_payment_orders_user (user_id, created_at),
    KEY idx_payment_orders_status_expiry (status, expires_at),
    CONSTRAINT fk_payment_orders_user FOREIGN KEY (user_id) REFERENCES users (id),
    CONSTRAINT fk_payment_orders_package FOREIGN KEY (package_id) REFERENCES packages (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS payment_refunds (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    refund_no VARCHAR(96) NOT NULL,
    order_id BIGINT UNSIGNED NOT NULL,
    amount_cents INT NOT NULL,
    reason VARCHAR(512) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    UNIQUE KEY uk_payment_refunds_no (refund_no),
    KEY idx_payment_refunds_order (order_id),
    CONSTRAINT fk_payment_refunds_order FOREIGN KEY (order_id) REFERENCES payment_orders (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
