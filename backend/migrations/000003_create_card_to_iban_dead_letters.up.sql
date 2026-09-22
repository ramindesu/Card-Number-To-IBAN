BEGIN;

CREATE TYPE card_to_iban_dead_letter_reason AS ENUM (
    'authentication_failed',
    'retries_exhausted'
);

CREATE TABLE card_to_iban_dead_letters (
    id UUID PRIMARY KEY,
    inquiry_id UUID NOT NULL UNIQUE,
    last_attempt_id UUID NOT NULL,
    reason card_to_iban_dead_letter_reason NOT NULL,
    attempt_count SMALLINT NOT NULL,
    last_meta_code INTEGER,
    last_error_type VARCHAR(64),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_card_to_iban_dead_letters_inquiry
        FOREIGN KEY (inquiry_id)
        REFERENCES card_to_iban_inquiries (request_id)
        ON DELETE CASCADE,
    CONSTRAINT fk_card_to_iban_dead_letters_last_attempt
        FOREIGN KEY (last_attempt_id)
        REFERENCES card_to_iban_upstream_attempts (id)
        ON DELETE CASCADE,
    CONSTRAINT chk_card_to_iban_dead_letters_attempt_count
        CHECK (attempt_count > 0)
);

CREATE INDEX idx_card_to_iban_dead_letters_reason_created_at
    ON card_to_iban_dead_letters (reason, created_at);

COMMENT ON TABLE card_to_iban_dead_letters IS
    'Terminal Card-to-IBAN jobs requiring operator attention; never stores the full card number.';

COMMIT;
