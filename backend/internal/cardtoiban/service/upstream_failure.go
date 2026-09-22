package service

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/client"
	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/model"
	cardrepository "github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/repository"
)

type upstreamFailureMapping struct {
	attemptStatus model.CardToIBANUpstreamAttemptStatus
	inquiryStatus model.CardToIBANInquiryStatus
	errorCode     CardToIBANErrorCode
	httpStatus    int
	message       string
}

func (s *CardToIBANService) completeUpstreamFailure(
	ctx context.Context,
	requestID uuid.UUID,
	attemptID uuid.UUID,
	inquiryStartedAt time.Time,
	attemptStartedAt time.Time,
	completedAt time.Time,
	upstreamErr error,
) error {
	clientError := normalizeClientError(upstreamErr)
	mapping := mapUpstreamFailure(clientError.Kind)

	if err := s.repository.CompleteUpstreamAttempt(
		ctx,
		cardrepository.CardToIBANUpstreamAttemptCompletion{
			ID:                 attemptID,
			Status:             mapping.attemptStatus,
			UpstreamHTTPStatus: clientError.HTTPStatus,
			MetaCode:           clientError.MetaCode,
			ErrorType:          clientError.ErrorType,
			Duration:           elapsed(attemptStartedAt, completedAt),
			ResponseJSON:       sanitizeProviderResponse(clientError.RawResponse),
			FinishedAt:         completedAt,
		},
	); err != nil {
		s.bestEffortCompleteInquiryAsInternalError(
			ctx,
			requestID,
			inquiryStartedAt,
			completedAt,
		)
		return newInternalError(requestID, err)
	}

	if err := s.repository.CompleteInquiry(
		ctx,
		cardrepository.CardToIBANInquiryCompletion{
			RequestID:          requestID,
			Status:             mapping.inquiryStatus,
			ResponseHTTPStatus: mapping.httpStatus,
			TotalDuration:      elapsed(inquiryStartedAt, completedAt),
			CompletedAt:        completedAt,
		},
	); err != nil {
		return newInternalError(requestID, err)
	}

	return &CardToIBANError{
		Code:       mapping.errorCode,
		Message:    mapping.message,
		HTTPStatus: mapping.httpStatus,
		RequestID:  requestID,
		Cause:      upstreamErr,
	}
}

func normalizeClientError(err error) *client.CardToIBANError {
	var clientError *client.CardToIBANError
	if errors.As(err, &clientError) {
		return clientError
	}

	return &client.CardToIBANError{
		Kind:  client.CardToIBANErrorKindTransport,
		Cause: err,
	}
}

func mapUpstreamFailure(kind client.CardToIBANErrorKind) upstreamFailureMapping {
	switch kind {
	case client.CardToIBANErrorKindRejected:
		return upstreamFailureMapping{
			attemptStatus: model.CardToIBANUpstreamAttemptStatusRejected,
			inquiryStatus: model.CardToIBANInquiryStatusUpstreamRejected,
			errorCode:     CardToIBANErrorCodeUpstreamRejected,
			httpStatus:    http.StatusUnprocessableEntity,
			message:       "درخواست توسط سرویس استعلام رد شد",
		}
	case client.CardToIBANErrorKindAuthentication:
		return upstreamFailureMapping{
			attemptStatus: model.CardToIBANUpstreamAttemptStatusAuthFailed,
			inquiryStatus: model.CardToIBANInquiryStatusUpstreamAuthFailed,
			errorCode:     CardToIBANErrorCodeUpstreamAuthFailed,
			httpStatus:    http.StatusBadGateway,
			message:       "احراز هویت سرویس استعلام ناموفق بود",
		}
	case client.CardToIBANErrorKindTimeout:
		return upstreamFailureMapping{
			attemptStatus: model.CardToIBANUpstreamAttemptStatusTimeout,
			inquiryStatus: model.CardToIBANInquiryStatusUpstreamTimeout,
			errorCode:     CardToIBANErrorCodeUpstreamTimeout,
			httpStatus:    http.StatusGatewayTimeout,
			message:       "مهلت پاسخ سرویس استعلام به پایان رسید",
		}
	case client.CardToIBANErrorKindUnavailable:
		return upstreamFailureMapping{
			attemptStatus: model.CardToIBANUpstreamAttemptStatusUnavailable,
			inquiryStatus: model.CardToIBANInquiryStatusUpstreamUnavailable,
			errorCode:     CardToIBANErrorCodeUpstreamUnavailable,
			httpStatus:    http.StatusServiceUnavailable,
			message:       "سرویس استعلام موقتاً در دسترس نیست",
		}
	default:
		return upstreamFailureMapping{
			attemptStatus: model.CardToIBANUpstreamAttemptStatusTransportError,
			inquiryStatus: model.CardToIBANInquiryStatusUpstreamUnavailable,
			errorCode:     CardToIBANErrorCodeUpstreamUnavailable,
			httpStatus:    http.StatusBadGateway,
			message:       "ارتباط با سرویس استعلام برقرار نشد",
		}
	}
}
