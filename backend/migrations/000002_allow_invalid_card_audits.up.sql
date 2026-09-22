BEGIN;

ALTER TABLE card_to_iban_inquiries
    DROP CONSTRAINT chk_card_to_iban_inquiries_card_masked;

COMMENT ON COLUMN card_to_iban_inquiries.card_masked IS
    'Masked card for valid-shaped input, or a safe label for invalid input.';

COMMIT;
