package service

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/client"
	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/model"
	cardrepository "github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/repository"
)

func (s *CardToIBANService) convertWithRetry(
	ctx context.Context,
	requestID uuid.UUID,
	inquiryStartedAt time.Time,
	normalizedCard string,
) (CardToIBANResponse, error) {
	for attemptNumber := 1; attemptNumber <= s.maxAttempts; attemptNumber++ {
		attemptID := s.idGenerator()
		trackID := s.idGenerator()
		attemptStartedAt := s.clock()
		attempt := model.CardToIBANUpstreamAttempt{
			ID:            attemptID,
			InquiryID:     requestID,
			AttemptNumber: attemptNumber,
			TrackID:       trackID,
			Status:        model.CardToIBANUpstreamAttemptStatusStarted,
			HandlingType:  s.handlingType,
			StartedAt:     attemptStartedAt,
		}

		if err := s.repository.CreateUpstreamAttempt(ctx, attempt); err != nil {
			completedAt := s.clock()
			s.bestEffortCompleteInquiryAsInternalError(
				ctx,
				requestID,
				inquiryStartedAt,
				completedAt,
			)
			return CardToIBANResponse{}, newInternalError(requestID, err)
		}

		var result client.CardToIBANResult
		executionErr := s.executor.Execute(ctx, func(executionContext context.Context) error {
			var clientErr error
			result, clientErr = s.client.Convert(
				executionContext,
				client.CardToIBANRequest{
					Card:         normalizedCard,
					TrackID:      trackID,
					HandlingType: s.handlingType,
				},
			)
			return clientErr
		})
		completedAt := s.clock()

		if executionErr == nil {
			return s.completeSuccess(
				ctx,
				requestID,
				attemptID,
				trackID,
				inquiryStartedAt,
				attemptStartedAt,
				completedAt,
				result,
			)
		}

		clientError := normalizeClientError(executionErr)
		mapping := mapUpstreamFailure(clientError.Kind)
		if err := s.completeUpstreamAttemptFailure(
			ctx,
			attemptID,
			attemptStartedAt,
			completedAt,
			clientError,
			mapping,
		); err != nil {
			s.bestEffortCompleteInquiryAsInternalError(
				ctx,
				requestID,
				inquiryStartedAt,
				completedAt,
			)
			return CardToIBANResponse{}, newInternalError(requestID, err)
		}

		if shouldRetry(clientError.Kind) && attemptNumber < s.maxAttempts {
			if err := s.sleeper(
				ctx,
				retryDelay(s.retryBaseDelay, attemptNumber),
			); err == nil {
				continue
			}
		}

		return CardToIBANResponse{}, s.completeFinalUpstreamFailure(
			ctx,
			requestID,
			attemptID,
			attemptNumber,
			inquiryStartedAt,
			completedAt,
			clientError,
			mapping,
		)
	}

	return CardToIBANResponse{}, newInternalError(
		requestID,
		context.Canceled,
	)
}

func (s *CardToIBANService) completeSuccess(
	ctx context.Context,
	requestID uuid.UUID,
	attemptID uuid.UUID,
	trackID uuid.UUID,
	inquiryStartedAt time.Time,
	attemptStartedAt time.Time,
	completedAt time.Time,
	result client.CardToIBANResult,
) (CardToIBANResponse, error) {
	upstreamHTTPStatus := result.HTTPStatus
	metaCode := result.MetaCode
	if err := s.repository.CompleteUpstreamAttempt(
		ctx,
		cardrepository.CardToIBANUpstreamAttemptCompletion{
			ID:                 attemptID,
			Status:             model.CardToIBANUpstreamAttemptStatusSucceeded,
			UpstreamHTTPStatus: &upstreamHTTPStatus,
			MetaCode:           &metaCode,
			Duration:           elapsed(attemptStartedAt, completedAt),
			ResponseJSON:       sanitizeProviderResponse(result.RawResponse),
			FinishedAt:         completedAt,
		},
	); err != nil {
		s.bestEffortCompleteInquiryAsInternalError(
			ctx,
			requestID,
			inquiryStartedAt,
			completedAt,
		)
		return CardToIBANResponse{}, newInternalError(requestID, err)
	}

	if err := s.repository.CompleteInquiry(
		ctx,
		cardrepository.CardToIBANInquiryCompletion{
			RequestID:          requestID,
			Status:             model.CardToIBANInquiryStatusSucceeded,
			ResponseHTTPStatus: http.StatusOK,
			TotalDuration:      elapsed(inquiryStartedAt, completedAt),
			CompletedAt:        completedAt,
		},
	); err != nil {
		return CardToIBANResponse{}, newInternalError(requestID, err)
	}

	return CardToIBANResponse{
		RequestID:          requestID,
		TrackID:            trackID,
		BankName:           result.Data.BankName,
		EnglishBankName:    result.Data.EnglishBankName,
		IBAN:               result.Data.IBAN,
		Deposit:            result.Data.Deposit,
		DepositDescription: result.Data.DepositDescription,
		DepositOwners:      result.Data.DepositOwners,
		DepositStatus:      result.Data.DepositStatus,
	}, nil
}

func shouldRetry(kind client.CardToIBANErrorKind) bool {
	switch kind {
	case client.CardToIBANErrorKindTimeout,
		client.CardToIBANErrorKindUnavailable,
		client.CardToIBANErrorKindTransport:
		return true
	default:
		return false
	}
}

func retryDelay(baseDelay time.Duration, completedAttempts int) time.Duration {
	delay := baseDelay
	for attempt := 1; attempt < completedAttempts; attempt++ {
		delay *= 2
	}
	return delay
}

func sleepWithContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
