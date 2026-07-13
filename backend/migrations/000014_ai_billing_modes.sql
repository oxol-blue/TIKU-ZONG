ALTER TABLE ai_models ADD COLUMN billing_mode VARCHAR(16) NOT NULL DEFAULT 'fixed' AFTER ai_charge_count;
ALTER TABLE ai_models ADD COLUMN token_unit INT NOT NULL DEFAULT 1000 AFTER billing_mode;
ALTER TABLE ai_models ADD COLUMN cost_per_million_tokens_cents INT NOT NULL DEFAULT 0 AFTER token_unit;
ALTER TABLE ai_models ADD COLUMN cost_markup_percent INT NOT NULL DEFAULT 0 AFTER cost_per_million_tokens_cents;
ALTER TABLE ai_models ADD COLUMN cost_unit_cents INT NOT NULL DEFAULT 1 AFTER cost_markup_percent;
