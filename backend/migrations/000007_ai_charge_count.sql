ALTER TABLE question_ai ADD COLUMN charge_count INT NOT NULL DEFAULT 1 AFTER token_count;
