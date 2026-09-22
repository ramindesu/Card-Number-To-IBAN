package service

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/cardnumber"
	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/client"
	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/model"
	cardrepository "github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/repository"
)

// CardToIBANServiceConfig contains the service dependencies and policies.
type CardToIBANServiceConfig struct {
	Repository     cardrepository.CardToIBANRepository
	Client         client.CardToIBANClient
	FingerprintKey []byte
	HandlingType   model.CardToIBANRequestHandlingType
	Clock          func() time.Time
	IDGenerator    func() uuid.UUID
}

// CardToIBANService orchestrates validation, audit persistence, and the
// external Card-to-IBAN call.
type CardToIBANService struct {
	repository     cardrepository.CardToIBANRepository
	client         client.CardToIBANClient
	fingerprintKey []byte
	handlingType   model.CardToIBANRequestHandlingType
	clock          func() time.Time
	idGenerator    func() uuid.UUID
}

// NewCardToIBANService constructs a service with explicit dependencies.
func NewCardToIBANService(config CardToIBANServiceConfig) (*CardToIBANService, error) {
	if config.Repository == nil {
		return nil, errors.New("card-to-IBAN repository is required")
	}

	if config.Client == nil {
		return nil, errors.New("card-to-IBAN client is required")
	}

	if _, err := cardnumber.Fingerprint("", config.FingerprintKey); err != nil {
		return nil, err
	}

	if config.HandlingType == "" {
		config.HandlingType = model.CardToIBANRequestHandlingTypeStandard
	}

	if !config.HandlingType.IsValid() {
		return nil, errors.New("invalid card-to-IBAN request handling type")
	}

	if config.Clock == nil {
		config.Clock = func() time.Time { return time.Now().UTC() }
	}

	if config.IDGenerator == nil {
		config.IDGenerator = uuid.New
	}

	return &CardToIBANService{
		repository:     config.Repository,
		client:         config.Client,
		fingerprintKey: append([]byte(nil), config.FingerprintKey...),
		handlingType:   config.HandlingType,
		clock:          config.Clock,
		idGenerator:    config.IDGenerator,
	}, nil
}

// Convert validates a card, records its audit lifecycle, and calls the
// external provider.
func (s *CardToIBANService) Convert(
	ctx context.Context,
	request CardToIBANRequest,
) (CardToIBANResponse, error) {
	startedAt := s.clock()
	requestID := s.idGenerator()
	normalizedCard := cardnumber.Normalize(request.Card)

	fingerprint, err := cardnumber.Fingerprint(normalizedCard, s.fingerprintKey)
	if err != nil {
		return CardToIBANResponse{}, newInternalError(requestID, err)
	}

	inquiry := model.CardToIBANInquiry{
		RequestID:       requestID,
		CardMasked:      cardnumber.Mask(normalizedCard),
		CardFingerprint: fingerprint,
		ClientIP:        request.ClientIP,
		Status:          model.CardToIBANInquiryStatusReceived,
		CreatedAt:       startedAt,
	}

	if err := s.repository.CreateInquiry(ctx, inquiry); err != nil {
		return CardToIBANResponse{}, newInternalError(requestID, err)
	}

	if err := cardnumber.Validate(normalizedCard); err != nil {
		return CardToIBANResponse{}, s.completeValidationFailure(
			ctx,
			requestID,
			startedAt,
			err,
		)
	}

	attemptID := s.idGenerator()
	trackID := s.idGenerator()
	attemptStartedAt := s.clock()
	attempt := model.CardToIBANUpstreamAttempt{
		ID:            attemptID,
		InquiryID:     requestID,
		AttemptNumber: 1,
		TrackID:       trackID,
		Status:        model.CardToIBANUpstreamAttemptStatusStarted,
		HandlingType:  s.handlingType,
		StartedAt:     attemptStartedAt,
	}

	if err := s.repository.CreateUpstreamAttempt(ctx, attempt); err != nil {
		completedAt := s.clock()
		s.bestEffortCompleteInquiryAsInternalError(ctx, requestID, startedAt, completedAt)
		return CardToIBANResponse{}, newInternalError(requestID, err)
	}

	result, clientErr := s.client.Convert(ctx, client.CardToIBANRequest{
		Card:         normalizedCard,
		TrackID:      trackID,
		HandlingType: s.handlingType,
	})
	completedAt := s.clock()

	if clientErr != nil {
		return CardToIBANResponse{}, s.completeUpstreamFailure(
			ctx,
			requestID,
			attemptID,
			startedAt,
			attemptStartedAt,
			completedAt,
			clientErr,
		)
	}

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
		s.bestEffortCompleteInquiryAsInternalError(ctx, requestID, startedAt, completedAt)
		return CardToIBANResponse{}, newInternalError(requestID, err)
	}

	if err := s.repository.CompleteInquiry(
		ctx,
		cardrepository.CardToIBANInquiryCompletion{
			RequestID:          requestID,
			Status:             model.CardToIBANInquiryStatusSucceeded,
			ResponseHTTPStatus: http.StatusOK,
			TotalDuration:      elapsed(startedAt, completedAt),
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

func (s *CardToIBANService) completeValidationFailure(
	ctx context.Context,
	requestID uuid.UUID,
	startedAt time.Time,
	validationErr error,
) error {
	completedAt := s.clock()
	if err := s.repository.CompleteInquiry(
		ctx,
		cardrepository.CardToIBANInquiryCompletion{
			RequestID:          requestID,
			Status:             model.CardToIBANInquiryStatusValidationFailed,
			ResponseHTTPStatus: http.StatusUnprocessableEntity,
			TotalDuration:      elapsed(startedAt, completedAt),
			CompletedAt:        completedAt,
		},
	); err != nil {
		return newInternalError(requestID, err)
	}

	var cardValidationError *cardnumber.ValidationError
	if !errors.As(validationErr, &cardValidationError) {
		return newInternalError(requestID, validationErr)
	}

	return &CardToIBANError{
		Code:           CardToIBANErrorCodeValidation,
		Message:        cardValidationError.Message,
		HTTPStatus:     http.StatusUnprocessableEntity,
		RequestID:      requestID,
		ValidationCode: cardValidationError.Code,
		Cause:          validationErr,
	}
}

func (s *CardToIBANService) bestEffortCompleteInquiryAsInternalError(
	ctx context.Context,
	requestID uuid.UUID,
	startedAt time.Time,
	completedAt time.Time,
) {
	_ = s.repository.CompleteInquiry(
		ctx,
		cardrepository.CardToIBANInquiryCompletion{
			RequestID:          requestID,
			Status:             model.CardToIBANInquiryStatusInternalError,
			ResponseHTTPStatus: http.StatusInternalServerError,
			TotalDuration:      elapsed(startedAt, completedAt),
			CompletedAt:        completedAt,
		},
	)
}

func newInternalError(requestID uuid.UUID, cause error) *CardToIBANError {
	return &CardToIBANError{
		Code:       CardToIBANErrorCodeInternal,
		Message:    "خطای داخلی رخ داده است",
		HTTPStatus: http.StatusInternalServerError,
		RequestID:  requestID,
		Cause:      cause,
	}
}

func elapsed(startedAt, completedAt time.Time) time.Duration {
	duration := completedAt.Sub(startedAt)
	if duration < 0 {
		return 0
	}

	return duration
}
