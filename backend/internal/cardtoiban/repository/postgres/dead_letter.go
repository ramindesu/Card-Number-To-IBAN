package postgres

import (
	"context"
	"fmt"

	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/model"
)

// CreateDeadLetter records a terminal operational failure without storing the
// original card number.
func (r *CardToIBANRepository) CreateDeadLetter(
	ctx context.Context,
	deadLetter model.CardToIBANDeadLetter,
) error {
	const query = `
		INSERT INTO card_to_iban_dead_letters (
			id,
			inquiry_id,
			last_attempt_id,
			reason,
			attempt_count,
			last_meta_code,
			last_error_type,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	if _, err := r.database.Exec(
		ctx,
		query,
		deadLetter.ID,
		deadLetter.InquiryID,
		deadLetter.LastAttemptID,
		deadLetter.Reason,
		deadLetter.AttemptCount,
		deadLetter.LastMetaCode,
		deadLetter.LastErrorType,
		deadLetter.CreatedAt,
	); err != nil {
		return fmt.Errorf("create card-to-IBAN dead letter: %w", err)
	}

	return nil
}
