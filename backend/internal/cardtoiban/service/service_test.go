package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/cardnumber"
	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/client"
	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/model"
	cardrepository "github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/repository"
	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/service"
)

func TestCardToIBANServiceValidationFailure(t *testing.T) {
	t.Parallel()

	requestID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	startedAt := time.Date(2026, time.September, 22, 10, 0, 0, 0, time.UTC)
	repository := &fakeCardToIBANRepository{}
	providerClient := &fakeCardToIBANClient{}

	applicationService := newTestService(
		t,
		repository,
		providerClient,
		sequenceClock(startedAt, startedAt.Add(5*time.Millisecond)),
		sequenceIDGenerator(requestID),
	)

	response, err := applicationService.Convert(context.Background(), service.CardToIBANRequest{
		Card: "۱۲۳۴",
	})
	require.Error(t, err)
	assert.Equal(t, service.CardToIBANResponse{}, response)

	var serviceError *service.CardToIBANError
	require.ErrorAs(t, err, &serviceError)
	assert.Equal(t, service.CardToIBANErrorCodeValidation, serviceError.Code)
	assert.Equal(t, cardnumber.ValidationErrorCodeInvalidLength, serviceError.ValidationCode)
	assert.Equal(t, http.StatusUnprocessableEntity, serviceError.HTTPStatus)
	assert.Equal(t, requestID, serviceError.RequestID)

	require.Len(t, repository.createdInquiries, 1)
	createdInquiry := repository.createdInquiries[0]
	assert.Equal(t, requestID, createdInquiry.RequestID)
	assert.Equal(t, "invalid", createdInquiry.CardMasked)
	assert.Len(t, createdInquiry.CardFingerprint, 64)
	assert.NotContains(t, createdInquiry.CardFingerprint, "1234")
	assert.Equal(t, model.CardToIBANInquiryStatusReceived, createdInquiry.Status)

	require.Len(t, repository.completedInquiries, 1)
	completedInquiry := repository.completedInquiries[0]
	assert.Equal(t, model.CardToIBANInquiryStatusValidationFailed, completedInquiry.Status)
	assert.Equal(t, http.StatusUnprocessableEntity, completedInquiry.ResponseHTTPStatus)
	assert.Equal(t, 5*time.Millisecond, completedInquiry.TotalDuration)

	assert.Empty(t, repository.createdAttempts)
	assert.Empty(t, providerClient.requests)
}

func TestCardToIBANServiceSuccess(t *testing.T) {
	t.Parallel()

	requestID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	attemptID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	trackID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	deadLetterID := uuid.MustParse("44444444-4444-4444-8444-444444444444")
	startedAt := time.Date(2026, time.September, 22, 10, 0, 0, 0, time.UTC)
	repository := &fakeCardToIBANRepository{}
	providerClient := &fakeCardToIBANClient{
		result: client.CardToIBANResult{
			HTTPStatus: http.StatusOK,
			MetaCode:   0,
			Data: client.CardToIBANData{
				BankName:        "بانک ملی ایران",
				EnglishBankName: "Bank Melli Iran",
				IBAN:            "IR820540102680020817909002",
				Deposit:         "0101234567001",
				DepositOwners:   "Test Owner",
				DepositStatus:   "active",
			},
			RawResponse: json.RawMessage(`{
				"data": {
					"card": "6037997512345670",
					"iban": "IR820540102680020817909002"
				},
				"meta": {
					"code": 0,
					"isSuccess": true
				}
			}`),
		},
	}

	applicationService := newTestService(
		t,
		repository,
		providerClient,
		sequenceClock(
			startedAt,
			startedAt.Add(10*time.Millisecond),
			startedAt.Add(760*time.Millisecond),
		),
		sequenceIDGenerator(requestID, attemptID, trackID, deadLetterID),
	)

	response, err := applicationService.Convert(context.Background(), service.CardToIBANRequest{
		Card: "۶۰۳۷-۹۹۷۵-۱۲۳۴-۵۶۷۰",
	})
	require.NoError(t, err)
	assert.Equal(t, requestID, response.RequestID)
	assert.Equal(t, trackID, response.TrackID)
	assert.Equal(t, "IR820540102680020817909002", response.IBAN)
	assert.Equal(t, "بانک ملی ایران", response.BankName)

	require.Len(t, providerClient.requests, 1)
	assert.Equal(t, "6037997512345670", providerClient.requests[0].Card)
	assert.Equal(t, trackID, providerClient.requests[0].TrackID)
	assert.Equal(
		t,
		model.CardToIBANRequestHandlingTypeStandard,
		providerClient.requests[0].HandlingType,
	)

	require.Len(t, repository.createdInquiries, 1)
	assert.Equal(t, "603799******5670", repository.createdInquiries[0].CardMasked)

	require.Len(t, repository.createdAttempts, 1)
	assert.Equal(t, attemptID, repository.createdAttempts[0].ID)
	assert.Equal(t, trackID, repository.createdAttempts[0].TrackID)

	require.Len(t, repository.completedAttempts, 1)
	completedAttempt := repository.completedAttempts[0]
	assert.Equal(t, model.CardToIBANUpstreamAttemptStatusSucceeded, completedAttempt.Status)
	assert.Equal(t, 750*time.Millisecond, completedAttempt.Duration)
	assert.NotContains(t, string(completedAttempt.ResponseJSON), "6037997512345670")
	assert.JSONEq(t, `{
		"data": {
			"card": "603799******5670",
			"iban": "IR820540102680020817909002"
		},
		"meta": {
			"code": 0,
			"isSuccess": true
		}
	}`, string(completedAttempt.ResponseJSON))

	require.Len(t, repository.completedInquiries, 1)
	assert.Equal(
		t,
		model.CardToIBANInquiryStatusSucceeded,
		repository.completedInquiries[0].Status,
	)
	assert.Equal(t, http.StatusOK, repository.completedInquiries[0].ResponseHTTPStatus)
	assert.Equal(t, 760*time.Millisecond, repository.completedInquiries[0].TotalDuration)
}

func TestCardToIBANServiceTimeout(t *testing.T) {
	t.Parallel()

	requestID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	attemptID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	trackID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	startedAt := time.Date(2026, time.September, 22, 10, 0, 0, 0, time.UTC)
	metaCode := 14
	errorType := "RequestTimeOut"
	repository := &fakeCardToIBANRepository{}
	providerClient := &fakeCardToIBANClient{
		err: &client.CardToIBANError{
			Kind:        client.CardToIBANErrorKindTimeout,
			MetaCode:    &metaCode,
			ErrorType:   &errorType,
			RawResponse: json.RawMessage(`{"data":null,"meta":{"code":14}}`),
			Cause:       context.DeadlineExceeded,
		},
	}

	applicationService := newTestService(
		t,
		repository,
		providerClient,
		sequenceClock(
			startedAt,
			startedAt.Add(10*time.Millisecond),
			startedAt.Add(2*time.Second),
		),
		sequenceIDGenerator(requestID, attemptID, trackID),
	)

	_, err := applicationService.Convert(context.Background(), service.CardToIBANRequest{
		Card: "6037997512345670",
	})
	require.Error(t, err)

	var serviceError *service.CardToIBANError
	require.ErrorAs(t, err, &serviceError)
	assert.Equal(t, service.CardToIBANErrorCodeUpstreamTimeout, serviceError.Code)
	assert.Equal(t, http.StatusGatewayTimeout, serviceError.HTTPStatus)

	require.Len(t, repository.completedAttempts, 1)
	assert.Equal(
		t,
		model.CardToIBANUpstreamAttemptStatusTimeout,
		repository.completedAttempts[0].Status,
	)
	assert.Equal(t, 1990*time.Millisecond, repository.completedAttempts[0].Duration)
	assert.Equal(t, &metaCode, repository.completedAttempts[0].MetaCode)

	require.Len(t, repository.completedInquiries, 1)
	assert.Equal(
		t,
		model.CardToIBANInquiryStatusUpstreamTimeout,
		repository.completedInquiries[0].Status,
	)
	assert.Equal(
		t,
		http.StatusGatewayTimeout,
		repository.completedInquiries[0].ResponseHTTPStatus,
	)

	require.Len(t, repository.deadLetters, 1)
	assert.Equal(t, deadLetterID, repository.deadLetters[0].ID)
	assert.Equal(t, requestID, repository.deadLetters[0].InquiryID)
	assert.Equal(t, attemptID, repository.deadLetters[0].LastAttemptID)
	assert.Equal(t, 1, repository.deadLetters[0].AttemptCount)
	assert.Equal(
		t,
		model.CardToIBANDeadLetterReasonRetriesExhausted,
		repository.deadLetters[0].Reason,
	)
}

func TestCardToIBANServiceCreateInquiryFailure(t *testing.T) {
	t.Parallel()

	requestID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	repositoryError := errors.New("database is unavailable")
	repository := &fakeCardToIBANRepository{createInquiryError: repositoryError}
	providerClient := &fakeCardToIBANClient{}

	applicationService := newTestService(
		t,
		repository,
		providerClient,
		sequenceClock(time.Now()),
		sequenceIDGenerator(requestID),
	)

	_, err := applicationService.Convert(context.Background(), service.CardToIBANRequest{
		Card: "6037997512345670",
	})
	require.Error(t, err)

	var serviceError *service.CardToIBANError
	require.ErrorAs(t, err, &serviceError)
	assert.Equal(t, service.CardToIBANErrorCodeInternal, serviceError.Code)
	assert.Equal(t, http.StatusInternalServerError, serviceError.HTTPStatus)
	assert.ErrorIs(t, err, repositoryError)
	assert.Empty(t, providerClient.requests)
	assert.Empty(t, repository.createdAttempts)
}

func newTestService(
	t *testing.T,
	repository cardrepository.CardToIBANRepository,
	providerClient client.CardToIBANClient,
	clock func() time.Time,
	idGenerator func() uuid.UUID,
) *service.CardToIBANService {
	t.Helper()

	applicationService, err := service.NewCardToIBANService(service.CardToIBANServiceConfig{
		Repository:     repository,
		Client:         providerClient,
		FingerprintKey: []byte("0123456789abcdef0123456789abcdef"),
		HandlingType:   model.CardToIBANRequestHandlingTypeStandard,
		Clock:          clock,
		IDGenerator:    idGenerator,
	})
	require.NoError(t, err)

	return applicationService
}

func sequenceClock(values ...time.Time) func() time.Time {
	index := 0
	return func() time.Time {
		value := values[index]
		index++
		return value
	}
}

func sequenceIDGenerator(values ...uuid.UUID) func() uuid.UUID {
	index := 0
	return func() uuid.UUID {
		value := values[index]
		index++
		return value
	}
}

type fakeCardToIBANRepository struct {
	createdInquiries   []model.CardToIBANInquiry
	completedInquiries []cardrepository.CardToIBANInquiryCompletion
	createdAttempts    []model.CardToIBANUpstreamAttempt
	completedAttempts  []cardrepository.CardToIBANUpstreamAttemptCompletion
	deadLetters        []model.CardToIBANDeadLetter

	createInquiryError         error
	completeInquiryError       error
	createUpstreamAttemptError error
	completeUpstreamError      error
	createDeadLetterError      error
}

func (r *fakeCardToIBANRepository) CreateInquiry(
	_ context.Context,
	inquiry model.CardToIBANInquiry,
) error {
	r.createdInquiries = append(r.createdInquiries, inquiry)
	return r.createInquiryError
}

func (r *fakeCardToIBANRepository) CompleteInquiry(
	_ context.Context,
	completion cardrepository.CardToIBANInquiryCompletion,
) error {
	r.completedInquiries = append(r.completedInquiries, completion)
	return r.completeInquiryError
}

func (r *fakeCardToIBANRepository) CreateUpstreamAttempt(
	_ context.Context,
	attempt model.CardToIBANUpstreamAttempt,
) error {
	r.createdAttempts = append(r.createdAttempts, attempt)
	return r.createUpstreamAttemptError
}

func (r *fakeCardToIBANRepository) CompleteUpstreamAttempt(
	_ context.Context,
	completion cardrepository.CardToIBANUpstreamAttemptCompletion,
) error {
	r.completedAttempts = append(r.completedAttempts, completion)
	return r.completeUpstreamError
}

func (r *fakeCardToIBANRepository) CreateDeadLetter(
	_ context.Context,
	deadLetter model.CardToIBANDeadLetter,
) error {
	r.deadLetters = append(r.deadLetters, deadLetter)
	return r.createDeadLetterError
}

type fakeCardToIBANClient struct {
	requests []client.CardToIBANRequest
	result   client.CardToIBANResult
	err      error
}

func (c *fakeCardToIBANClient) Convert(
	_ context.Context,
	request client.CardToIBANRequest,
) (client.CardToIBANResult, error) {
	c.requests = append(c.requests, request)
	return c.result, c.err
}
