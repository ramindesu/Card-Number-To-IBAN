package model

import (
	"time"

	"github.com/google/uuid"
)

// CardToIBANDeadLetterReason explains why an inquiry needs operator attention.
type CardToIBANDeadLetterReason string

const (
	CardToIBANDeadLetterReasonAuthenticationFailed CardToIBANDeadLetterReason = "authentication_failed"
	CardToIBANDeadLetterReasonRetriesExhausted     CardToIBANDeadLetterReason = "retries_exhausted"
)

// CardToIBANDeadLetter stores safe metadata only; the original PAN is never
// persisted and replay therefore requires a new client submission.
type CardToIBANDeadLetter struct {
	ID            uuid.UUID
	InquiryID     uuid.UUID
	LastAttemptID uuid.UUID
	Reason        CardToIBANDeadLetterReason
	AttemptCount  int
	LastMetaCode  *int
	LastErrorType *string
	CreatedAt     time.Time
}
