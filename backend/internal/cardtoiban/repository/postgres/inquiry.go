package postgres

import (
	"context"
	"fmt"

	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/model"
	cardrepository "github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/repository"
)

const createCardToIBANInquiryQuery = `
INSERT INTO card_to_iban_inquiries (
    request_id,
    card_masked,
    card_fingerprint,
    client_ip,
    status,
    response_http_status,
    total_duration_ms,
    created_at,
    completed_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

const completeCardToIBANInquiryQuery = `
UPDATE card_to_iban_inquiries
SET
    status = $2,
    response_http_status = $3,
    total_duration_ms = $4,
    completed_at = $5
WHERE request_id = $1`

// CreateInquiry inserts the initial audit record before validation or any
// upstream call is performed.
func (r *CardToIBANRepository) CreateInquiry(
	ctx context.Context,
	inquiry model.CardToIBANInquiry,
) error {
	_, err := r.database.Exec(
		ctx,
		createCardToIBANInquiryQuery,
		inquiry.RequestID,
		inquiry.CardMasked,
		inquiry.CardFingerprint,
		inquiry.ClientIP,
		inquiry.Status,
		inquiry.ResponseHTTPStatus,
		nullableDurationMilliseconds(inquiry.TotalDuration),
		inquiry.CreatedAt,
		inquiry.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("create card-to-IBAN inquiry: %w", err)
	}

	return nil
}

// CompleteInquiry updates the final state of an existing inquiry.
func (r *CardToIBANRepository) CompleteInquiry(
	ctx context.Context,
	completion cardrepository.CardToIBANInquiryCompletion,
) error {
	commandTag, err := r.database.Exec(
		ctx,
		completeCardToIBANInquiryQuery,
		completion.RequestID,
		completion.Status,
		completion.ResponseHTTPStatus,
		completion.TotalDuration.Milliseconds(),
		completion.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("complete card-to-IBAN inquiry: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return cardrepository.ErrCardToIBANInquiryNotFound
	}

	return nil
}
