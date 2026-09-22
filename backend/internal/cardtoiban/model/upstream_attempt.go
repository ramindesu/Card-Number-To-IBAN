package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// CardToIBANUpstreamAttempt contains the audit information for one call to
// ZarinHub. One inquiry may have no attempt, one attempt, or multiple attempts.
type CardToIBANUpstreamAttempt struct {
	ID                 uuid.UUID
	InquiryID          uuid.UUID
	AttemptNumber      int
	TrackID            uuid.UUID
	Status             CardToIBANUpstreamAttemptStatus
	HandlingType       CardToIBANRequestHandlingType
	UpstreamHTTPStatus *int
	MetaCode           *int
	ErrorType          *string
	Duration           *time.Duration
	ResponseJSON       json.RawMessage
	StartedAt          time.Time
	FinishedAt         *time.Time

}
