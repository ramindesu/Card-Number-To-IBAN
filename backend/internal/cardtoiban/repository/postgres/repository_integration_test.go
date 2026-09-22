//go:build integration

package postgres_test

import (
	"context"
	"encoding/json"
	"net/netip"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/model"
	cardrepository "github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/repository"
	cardpostgres "github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/repository/postgres"
	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/platform/config"
	platformdatabase "github.com/ramindesu/Card-Number-To-IBAN/backend/internal/platform/database"
)

func TestCardToIBANRepositoryLifecycle(t *testing.T) {
	ctx := context.Background()

	databaseConfig, err := config.LoadDatabase()
	require.NoError(t, err)

	pool, err := platformdatabase.NewPostgresPool(ctx, databaseConfig)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	transaction, err := pool.Begin(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, transaction.Rollback(ctx))
	})

	repository := cardpostgres.NewCardToIBANRepository(transaction)
	now := time.Now().UTC().Truncate(time.Microsecond)
	clientIP := netip.MustParseAddr("127.0.0.1")
	requestID := uuid.New()

	inquiry := model.CardToIBANInquiry{
		RequestID:       requestID,
		CardMasked:      "603799******5678",
		CardFingerprint: "67f9cc880ae1823bc6b6fbbf82b262363573751f5b8a28d39604f27b5665ea8f",
		ClientIP:        &clientIP,
		Status:          model.CardToIBANInquiryStatusReceived,
		CreatedAt:       now,
	}

	require.NoError(t, repository.CreateInquiry(ctx, inquiry))

	attemptID := uuid.New()
	attempt := model.CardToIBANUpstreamAttempt{
		ID:            attemptID,
		InquiryID:     requestID,
		AttemptNumber: 1,
		TrackID:       uuid.New(),
		Status:        model.CardToIBANUpstreamAttemptStatusStarted,
		HandlingType:  model.CardToIBANRequestHandlingTypeStandard,
		StartedAt:     now.Add(time.Millisecond),
	}

	require.NoError(t, repository.CreateUpstreamAttempt(ctx, attempt))

	upstreamHTTPStatus := 200
	metaCode := 0
	responseJSON := json.RawMessage(`{
		"data": {
			"card": "603799******5678",
			"iban": "IR820540102680020817909002"
		},
		"meta": {
			"code": 0,
			"isSuccess": true
		}
	}`)

	require.NoError(t, repository.CompleteUpstreamAttempt(
		ctx,
		cardrepository.CardToIBANUpstreamAttemptCompletion{
			ID:                 attemptID,
			Status:             model.CardToIBANUpstreamAttemptStatusSucceeded,
			UpstreamHTTPStatus: &upstreamHTTPStatus,
			MetaCode:           &metaCode,
			Duration:           750 * time.Millisecond,
			ResponseJSON:       responseJSON,
			FinishedAt:         now.Add(751 * time.Millisecond),
		},
	))

	require.NoError(t, repository.CompleteInquiry(
		ctx,
		cardrepository.CardToIBANInquiryCompletion{
			RequestID:          requestID,
			Status:             model.CardToIBANInquiryStatusSucceeded,
			ResponseHTTPStatus: 200,
			TotalDuration:      800 * time.Millisecond,
			CompletedAt:        now.Add(800 * time.Millisecond),
		},
	))

	var (
		inquiryStatus      string
		responseHTTPStatus int
		totalDurationMS    int64
	)

	err = transaction.QueryRow(
		ctx,
		`SELECT status, response_http_status, total_duration_ms
		 FROM card_to_iban_inquiries
		 WHERE request_id = $1`,
		requestID,
	).Scan(&inquiryStatus, &responseHTTPStatus, &totalDurationMS)
	require.NoError(t, err)
	assert.Equal(t, "succeeded", inquiryStatus)
	assert.Equal(t, 200, responseHTTPStatus)
	assert.EqualValues(t, 800, totalDurationMS)

	var (
		attemptStatus    string
		storedMetaCode   int
		storedDurationMS int64
		storedResponse   []byte
	)

	err = transaction.QueryRow(
		ctx,
		`SELECT status, meta_code, duration_ms, response_json
		 FROM card_to_iban_upstream_attempts
		 WHERE id = $1`,
		attemptID,
	).Scan(&attemptStatus, &storedMetaCode, &storedDurationMS, &storedResponse)
	require.NoError(t, err)
	assert.Equal(t, "succeeded", attemptStatus)
	assert.Equal(t, 0, storedMetaCode)
	assert.EqualValues(t, 750, storedDurationMS)
	assert.JSONEq(t, string(responseJSON), string(storedResponse))
}
