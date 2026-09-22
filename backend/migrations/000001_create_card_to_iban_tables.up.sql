BEGIN;

CREATE TYPE card_to_iban_inquiry_status AS ENUM (
    'received',
    'validation_failed',
    'succeeded',
    'upstream_rejected',
    'upstream_auth_failed',
    'upstream_timeout',
    'upstream_unavailable',
    'internal_error'
);

CREATE TYPE card_to_iban_upstream_attempt_status AS ENUM (
    'started',
    'succeeded',
    'rejected',
    'auth_failed',
    'timeout',
    'unavailable',
    'transport_error'
);

CREATE TYPE card_to_iban_request_handling_type AS ENUM (
    'standard',
    'advanced'
);

CREATE TABLE card_to_iban_inquiries (
    request_id UUID PRIMARY KEY,
    card_masked VARCHAR(19) NOT NULL,
    card_fingerprint CHAR(64) NOT NULL,
    client_ip INET,
    status card_to_iban_inquiry_status NOT NULL DEFAULT 'received',
    response_http_status SMALLINT,
    total_duration_ms BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,

    CONSTRAINT chk_card_to_iban_inquiries_card_masked
        CHECK (card_masked ~ '^[0-9]{6}\*{6}[0-9]{4}$'),
    CONSTRAINT chk_card_to_iban_inquiries_card_fingerprint
        CHECK (card_fingerprint ~ '^[0-9a-f]{64}$'),
    CONSTRAINT chk_card_to_iban_inquiries_response_http_status
        CHECK (response_http_status IS NULL OR response_http_status BETWEEN 100 AND 599),
    CONSTRAINT chk_card_to_iban_inquiries_total_duration
        CHECK (total_duration_ms IS NULL OR total_duration_ms >= 0)
);

CREATE INDEX idx_card_to_iban_inquiries_card_fingerprint
    ON card_to_iban_inquiries (card_fingerprint);

CREATE INDEX idx_card_to_iban_inquiries_created_at
    ON card_to_iban_inquiries (created_at);

CREATE INDEX idx_card_to_iban_inquiries_status_created_at
    ON card_to_iban_inquiries (status, created_at);

CREATE TABLE card_to_iban_upstream_attempts (
    id UUID PRIMARY KEY,
    inquiry_id UUID NOT NULL,
    attempt_no SMALLINT NOT NULL,
    track_id UUID NOT NULL UNIQUE,
    status card_to_iban_upstream_attempt_status NOT NULL DEFAULT 'started',
    handling_type card_to_iban_request_handling_type NOT NULL DEFAULT 'standard',
    upstream_http_status SMALLINT,
    meta_code INTEGER,
    error_type VARCHAR(64),
    duration_ms BIGINT,
    response_json JSONB,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMPTZ,

    CONSTRAINT fk_card_to_iban_upstream_attempts_inquiry
        FOREIGN KEY (inquiry_id)
        REFERENCES card_to_iban_inquiries (request_id)
        ON DELETE CASCADE,
    CONSTRAINT uq_card_to_iban_upstream_attempts_inquiry_attempt
        UNIQUE (inquiry_id, attempt_no),
    CONSTRAINT chk_card_to_iban_upstream_attempts_attempt_no
        CHECK (attempt_no > 0),
    CONSTRAINT chk_card_to_iban_upstream_attempts_http_status
        CHECK (upstream_http_status IS NULL OR upstream_http_status BETWEEN 100 AND 599),
    CONSTRAINT chk_card_to_iban_upstream_attempts_duration
        CHECK (duration_ms IS NULL OR duration_ms >= 0)
);

CREATE INDEX idx_card_to_iban_upstream_attempts_status
    ON card_to_iban_upstream_attempts (status);

CREATE INDEX idx_card_to_iban_upstream_attempts_meta_code
    ON card_to_iban_upstream_attempts (meta_code);

CREATE INDEX idx_card_to_iban_upstream_attempts_started_at
    ON card_to_iban_upstream_attempts (started_at);

COMMENT ON TABLE card_to_iban_inquiries IS
    'Audit record for every request received by the Card-to-IBAN API.';

COMMENT ON TABLE card_to_iban_upstream_attempts IS
    'Audit record for every Card-to-IBAN call made to ZarinHub.';

COMMENT ON COLUMN card_to_iban_upstream_attempts.response_json IS
    'Complete ZarinHub response after masking the data.card value.';

COMMIT;
