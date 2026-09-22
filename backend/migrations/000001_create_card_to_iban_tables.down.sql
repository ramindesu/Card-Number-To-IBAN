BEGIN;

DROP TABLE IF EXISTS card_to_iban_upstream_attempts;
DROP TABLE IF EXISTS card_to_iban_inquiries;

DROP TYPE IF EXISTS card_to_iban_request_handling_type;
DROP TYPE IF EXISTS card_to_iban_upstream_attempt_status;
DROP TYPE IF EXISTS card_to_iban_inquiry_status;

COMMIT;
