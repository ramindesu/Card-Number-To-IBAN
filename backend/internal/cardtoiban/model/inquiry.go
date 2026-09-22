package model

import (
	"net/netip"
	"time"

	"github.com/google/uuid"
)

// CardToIBANInquiry contains the audit information for one request received by
// the Card-to-IBAN API. An inquiry exists even when validation fails before an
// upstream request is made.
type CardToIBANInquiry struct {
	RequestID          uuid.UUID
	CardMasked         string
	CardFingerprint    string
	ClientIP           *netip.Addr
	Status             CardToIBANInquiryStatus
	ResponseHTTPStatus *int
	TotalDuration      *time.Duration
	CreatedAt          time.Time
	CompletedAt        *time.Time
}
