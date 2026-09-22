package model

// CardToIBANInquiryStatus is the lifecycle state of a request received by the
// Card-to-IBAN API.
type CardToIBANInquiryStatus string

const (
	CardToIBANInquiryStatusReceived            CardToIBANInquiryStatus = "received"
	CardToIBANInquiryStatusValidationFailed    CardToIBANInquiryStatus = "validation_failed"
	CardToIBANInquiryStatusSucceeded           CardToIBANInquiryStatus = "succeeded"
	CardToIBANInquiryStatusUpstreamRejected    CardToIBANInquiryStatus = "upstream_rejected"
	CardToIBANInquiryStatusUpstreamAuthFailed  CardToIBANInquiryStatus = "upstream_auth_failed"
	CardToIBANInquiryStatusUpstreamTimeout     CardToIBANInquiryStatus = "upstream_timeout"
	CardToIBANInquiryStatusUpstreamUnavailable CardToIBANInquiryStatus = "upstream_unavailable"
	CardToIBANInquiryStatusInternalError       CardToIBANInquiryStatus = "internal_error"
)

// IsValid reports whether the inquiry status is supported by the domain.
func (s CardToIBANInquiryStatus) IsValid() bool {
	switch s {
	case CardToIBANInquiryStatusReceived,
		CardToIBANInquiryStatusValidationFailed,
		CardToIBANInquiryStatusSucceeded,
		CardToIBANInquiryStatusUpstreamRejected,
		CardToIBANInquiryStatusUpstreamAuthFailed,
		CardToIBANInquiryStatusUpstreamTimeout,
		CardToIBANInquiryStatusUpstreamUnavailable,
		CardToIBANInquiryStatusInternalError:
		return true
	default:
		return false
	}
}

// CardToIBANUpstreamAttemptStatus is the lifecycle state of one call to
// ZarinHub.
type CardToIBANUpstreamAttemptStatus string

const (
	CardToIBANUpstreamAttemptStatusStarted        CardToIBANUpstreamAttemptStatus = "started"
	CardToIBANUpstreamAttemptStatusSucceeded      CardToIBANUpstreamAttemptStatus = "succeeded"
	CardToIBANUpstreamAttemptStatusRejected       CardToIBANUpstreamAttemptStatus = "rejected"
	CardToIBANUpstreamAttemptStatusAuthFailed     CardToIBANUpstreamAttemptStatus = "auth_failed"
	CardToIBANUpstreamAttemptStatusTimeout        CardToIBANUpstreamAttemptStatus = "timeout"
	CardToIBANUpstreamAttemptStatusUnavailable    CardToIBANUpstreamAttemptStatus = "unavailable"
	CardToIBANUpstreamAttemptStatusTransportError CardToIBANUpstreamAttemptStatus = "transport_error"
)

// IsValid reports whether the upstream-attempt status is supported by the
// domain.
func (s CardToIBANUpstreamAttemptStatus) IsValid() bool {
	switch s {
	case CardToIBANUpstreamAttemptStatusStarted,
		CardToIBANUpstreamAttemptStatusSucceeded,
		CardToIBANUpstreamAttemptStatusRejected,
		CardToIBANUpstreamAttemptStatusAuthFailed,
		CardToIBANUpstreamAttemptStatusTimeout,
		CardToIBANUpstreamAttemptStatusUnavailable,
		CardToIBANUpstreamAttemptStatusTransportError:
		return true
	default:
		return false
	}
}

// CardToIBANRequestHandlingType controls how ZarinHub selects upstream
// providers for a request.
type CardToIBANRequestHandlingType string

const (
	CardToIBANRequestHandlingTypeStandard CardToIBANRequestHandlingType = "standard"
	CardToIBANRequestHandlingTypeAdvanced CardToIBANRequestHandlingType = "advanced"
)

// IsValid reports whether the request handling type is supported by ZarinHub.
func (t CardToIBANRequestHandlingType) IsValid() bool {
	switch t {
	case CardToIBANRequestHandlingTypeStandard,
		CardToIBANRequestHandlingTypeAdvanced:
		return true
	default:
		return false
	}
}
