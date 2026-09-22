BEGIN;

ALTER TABLE card_to_iban_inquiries
    ADD CONSTRAINT chk_card_to_iban_inquiries_card_masked
    CHECK (card_masked ~ '^[0-9]{6}\*{6}[0-9]{4}$');

COMMENT ON COLUMN card_to_iban_inquiries.card_masked IS NULL;

COMMIT;
