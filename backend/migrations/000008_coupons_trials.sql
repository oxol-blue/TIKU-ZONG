ALTER TABLE packages ADD COLUMN is_trial TINYINT NOT NULL DEFAULT 0 AFTER limit_count;
ALTER TABLE packages ADD COLUMN is_free TINYINT NOT NULL DEFAULT 0 AFTER is_trial;

CREATE TABLE IF NOT EXISTS coupons (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(64) NOT NULL,
    discount_type VARCHAR(16) NOT NULL,
    discount_value INT NOT NULL,
    total_limit INT NOT NULL DEFAULT 0,
    used_count INT NOT NULL DEFAULT 0,
    reserved_count INT NOT NULL DEFAULT 0,
    expires_at DATETIME(6) NULL,
    status TINYINT NOT NULL DEFAULT 1,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    UNIQUE KEY uk_coupons_code (code),
    KEY idx_coupons_status_expiry (status, expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

ALTER TABLE payment_orders ADD COLUMN coupon_id BIGINT UNSIGNED NULL AFTER provider;
ALTER TABLE payment_orders ADD COLUMN coupon_code VARCHAR(64) NOT NULL DEFAULT '' AFTER coupon_id;
ALTER TABLE payment_orders ADD COLUMN discount_cents INT NOT NULL DEFAULT 0 AFTER payable_cents;
ALTER TABLE payment_orders ADD KEY idx_payment_orders_coupon (coupon_id);
ALTER TABLE payment_orders ADD CONSTRAINT fk_payment_orders_coupon FOREIGN KEY (coupon_id) REFERENCES coupons (id);

CREATE TABLE IF NOT EXISTS coupon_reservations (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    coupon_id BIGINT UNSIGNED NOT NULL,
    order_id BIGINT UNSIGNED NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'reserved',
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    UNIQUE KEY uk_coupon_reservations_order (order_id),
    KEY idx_coupon_reservations_coupon (coupon_id, status),
    CONSTRAINT fk_coupon_reservations_coupon FOREIGN KEY (coupon_id) REFERENCES coupons (id),
    CONSTRAINT fk_coupon_reservations_order FOREIGN KEY (order_id) REFERENCES payment_orders (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
