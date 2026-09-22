package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/model"
)

var (
	// ErrCardToIBANInquiryNotFound is returned when an inquiry completion targets
	// a request that does not exist.
	ErrCardToIBANInquiryNotFound = errors.New("card-to-IBAN inquiry not found")

	// ErrCardToIBANUpstreamAttemptNotFound is returned when an attempt completion
	// targets an upstream attempt that does not exist.
	ErrCardToIBANUpstreamAttemptNotFound = errors.New("card-to-IBAN upstream attempt not found")
)

// CardToIBANInquiryCompletion contains the final mutable fields of an inquiry.
type CardToIBANInquiryCompletion struct {
	RequestID          uuid.UUID
	Status             model.CardToIBANInquiryStatus
	ResponseHTTPStatus int
	TotalDuration      time.Duration
	CompletedAt        time.Time
}

// CardToIBANUpstreamAttemptCompletion contains the final mutable fields of an
// upstream attempt.
type CardToIBANUpstreamAttemptCompletion struct {
	ID                 uuid.UUID
	Status             model.CardToIBANUpstreamAttemptStatus
	UpstreamHTTPStatus *int
	MetaCode           *int
	ErrorType          *string
	Duration           time.Duration
	ResponseJSON       json.RawMessage
	FinishedAt         time.Time
}

// CardToIBANRepository persists the audit lifecycle of inquiries and their
// ZarinHub attempts.
type CardToIBANRepository interface {
	CreateInquiry(context.Context, model.CardToIBANInquiry) error
	CompleteInquiry(context.Context, CardToIBANInquiryCompletion) error
	CreateUpstreamAttempt(context.Context, model.CardToIBANUpstreamAttempt) error
	CompleteUpstreamAttempt(context.Context, CardToIBANUpstreamAttemptCompletion) error
	CreateDeadLetter(context.Context, model.CardToIBANDeadLetter) error
}
