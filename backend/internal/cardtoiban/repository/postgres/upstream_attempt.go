package postgres

import (
	"context"
	"fmt"

	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/model"
	cardrepository "github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/repository"
)

const createCardToIBANUpstreamAttemptQuery = `
INSERT INTO card_to_iban_upstream_attempts (
    id,
    inquiry_id,
    attempt_no,
    track_id,
    status,
    handling_type,
    upstream_http_status,
    meta_code,
    error_type,
    duration_ms,
    response_json,
    started_at,
    finished_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

const completeCardToIBANUpstreamAttemptQuery = `
UPDATE card_to_iban_upstream_attempts
SET
    status = $2,
    upstream_http_status = $3,
    meta_code = $4,
    error_type = $5,
    duration_ms = $6,
    response_json = $7,
    finished_at = $8
WHERE id = $1`

// CreateUpstreamAttempt inserts an audit record immediately before calling
// ZarinHub.
func (r *CardToIBANRepository) CreateUpstreamAttempt(
	ctx context.Context,
	attempt model.CardToIBANUpstreamAttempt,
) error {
	_, err := r.database.Exec(
		ctx,
		createCardToIBANUpstreamAttemptQuery,
		attempt.ID,
		attempt.InquiryID,
		attempt.AttemptNumber,
		attempt.TrackID,
		attempt.Status,
		attempt.HandlingType,
		attempt.UpstreamHTTPStatus,
		attempt.MetaCode,
		attempt.ErrorType,
		nullableDurationMilliseconds(attempt.Duration),
		attempt.ResponseJSON,
		attempt.StartedAt,
		attempt.FinishedAt,
	)
	if err != nil {
		return fmt.Errorf("create card-to-IBAN upstream attempt: %w", err)
	}

	return nil
}

// CompleteUpstreamAttempt updates the final state of an existing upstream
// attempt.
func (r *CardToIBANRepository) CompleteUpstreamAttempt(
	ctx context.Context,
	completion cardrepository.CardToIBANUpstreamAttemptCompletion,
) error {
	commandTag, err := r.database.Exec(
		ctx,
		completeCardToIBANUpstreamAttemptQuery,
		completion.ID,
		completion.Status,
		completion.UpstreamHTTPStatus,
		completion.MetaCode,
		completion.ErrorType,
		completion.Duration.Milliseconds(),
		completion.ResponseJSON,
		completion.FinishedAt,
	)
	if err != nil {
		return fmt.Errorf("complete card-to-IBAN upstream attempt: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return cardrepository.ErrCardToIBANUpstreamAttemptNotFound
	}

	return nil
}
