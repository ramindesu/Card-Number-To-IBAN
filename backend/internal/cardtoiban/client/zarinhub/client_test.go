package zarinhub

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cardclient "github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/client"
	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/model"
)

func TestCardToIBANClientConvertCachesTokenAndMapsResponse(t *testing.T) {
	t.Parallel()

	const card = "6037997512345670"
	trackID := uuid.MustParse("b19e297d-0e04-49c4-8a9b-24b801e28d91")
	var authenticationCalls atomic.Int32
	var conversionCalls atomic.Int32

	handler := http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		assert.Equal(t, http.MethodPost, request.Method)
		assert.Equal(t, "application/json", request.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", request.Header.Get("Accept"))

		switch request.URL.Path {
		case authenticationPath:
			authenticationCalls.Add(1)

			var payload tokenRequest
			assert.NoError(t, json.NewDecoder(request.Body).Decode(&payload))
			assert.Equal(t, "api-user", payload.Username)
			assert.Equal(t, "api-password", payload.Password)
			assert.Empty(t, payload.RefreshToken)

			writeJSON(t, writer, http.StatusOK, map[string]any{
				"data": map[string]any{
					"accessToken":  "access-token",
					"tokenType":    "Bearer",
					"expiresTn":    3600,
					"refreshToken": "refresh-token",
				},
				"meta": map[string]any{"isSuccess": true, "code": 0},
			})

		case cardToIBANPath:
			conversionCalls.Add(1)
			assert.Equal(t, "Bearer access-token", request.Header.Get("Authorization"))

			var payload cardToIBANRequest
			assert.NoError(t, json.NewDecoder(request.Body).Decode(&payload))
			assert.Equal(t, card, payload.Card)
			assert.Equal(t, trackID, payload.TrackID)
			assert.Equal(t, "standard", payload.RequestHandlingType)

			writeJSON(t, writer, http.StatusOK, map[string]any{
				"data": map[string]any{
					"bankName":           "بانک ملی",
					"englishBankName":    "Melli",
					"card":               card,
					"deposit":            "0123456789",
					"depositDescription": "حساب کوتاه مدت",
					"depositOwners":      "کاربر تست",
					"depositStatus":      "active",
					"iban":               "IR820170000000123456789001",
				},
				"meta": map[string]any{
					"trackId":   trackID.String(),
					"status":    "done",
					"isSuccess": true,
					"code":      0,
				},
			})

		default:
			http.NotFound(writer, request)
		}
	})

	client := newTestClient(
		t,
		"https://zarin-hub.test",
		handlerHTTPClient(handler),
		nil,
	)
	request := cardclient.CardToIBANRequest{
		Card:         card,
		TrackID:      trackID,
		HandlingType: model.CardToIBANRequestHandlingTypeStandard,
	}

	for range 2 {
		result, err := client.Convert(context.Background(), request)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, result.HTTPStatus)
		assert.Equal(t, 0, result.MetaCode)
		assert.Equal(t, "بانک ملی", result.Data.BankName)
		assert.Equal(t, "Melli", result.Data.EnglishBankName)
		assert.Equal(t, "IR820170000000123456789001", result.Data.IBAN)
		assert.Equal(t, "0123456789", result.Data.Deposit)
		assert.JSONEq(t, `{
			"data": {
				"bankName": "بانک ملی",
				"englishBankName": "Melli",
				"card": "6037997512345670",
				"deposit": "0123456789",
				"depositDescription": "حساب کوتاه مدت",
				"depositOwners": "کاربر تست",
				"depositStatus": "active",
				"iban": "IR820170000000123456789001"
			},
			"meta": {
				"trackId": "b19e297d-0e04-49c4-8a9b-24b801e28d91",
				"status": "done",
				"isSuccess": true,
				"code": 0
			}
		}`, string(result.RawResponse))
	}

	assert.EqualValues(t, 1, authenticationCalls.Load())
	assert.EqualValues(t, 2, conversionCalls.Load())
}

func TestCardToIBANClientRefreshesExpiredToken(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.September, 22, 12, 0, 0, 0, time.UTC)
	var clockMutex sync.Mutex
	clock := func() time.Time {
		clockMutex.Lock()
		defer clockMutex.Unlock()
		return now
	}

	var handlerMutex sync.Mutex
	authenticationPayloads := make([]tokenRequest, 0, 2)
	authorizationHeaders := make([]string, 0, 2)

	handler := http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		handlerMutex.Lock()
		defer handlerMutex.Unlock()

		switch request.URL.Path {
		case authenticationPath:
			var payload tokenRequest
			assert.NoError(t, json.NewDecoder(request.Body).Decode(&payload))
			authenticationPayloads = append(authenticationPayloads, payload)
			callNumber := len(authenticationPayloads)

			writeJSON(t, writer, http.StatusOK, map[string]any{
				"data": map[string]any{
					"accessToken":  "access-token-" + string(rune('0'+callNumber)),
					"tokenType":    "Bearer",
					"expiresTn":    120,
					"refreshToken": "refresh-token-" + string(rune('0'+callNumber)),
				},
				"meta": map[string]any{"isSuccess": true, "code": 0},
			})

		case cardToIBANPath:
			authorizationHeaders = append(
				authorizationHeaders,
				request.Header.Get("Authorization"),
			)
			writeSuccessfulCardResponse(t, writer)

		default:
			http.NotFound(writer, request)
		}
	})

	client := newTestClient(
		t,
		"https://zarin-hub.test",
		handlerHTTPClient(handler),
		clock,
	)
	request := validClientRequest()

	_, err := client.Convert(context.Background(), request)
	require.NoError(t, err)

	clockMutex.Lock()
	now = now.Add(91 * time.Second)
	clockMutex.Unlock()

	_, err = client.Convert(context.Background(), request)
	require.NoError(t, err)

	handlerMutex.Lock()
	defer handlerMutex.Unlock()
	require.Len(t, authenticationPayloads, 2)
	assert.Equal(t, tokenRequest{
		Username: "api-user",
		Password: "api-password",
	}, authenticationPayloads[0])
	assert.Equal(t, tokenRequest{
		RefreshToken: "refresh-token-1",
	}, authenticationPayloads[1])
	assert.Equal(t, []string{
		"Bearer access-token-1",
		"Bearer access-token-2",
	}, authorizationHeaders)
}

func TestCardToIBANClientClassifiesProviderUnavailableResponse(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		switch request.URL.Path {
		case authenticationPath:
			writeTokenResponse(t, writer)
		case cardToIBANPath:
			writeJSON(t, writer, http.StatusServiceUnavailable, map[string]any{
				"data": nil,
				"meta": map[string]any{
					"isSuccess":    false,
					"code":         7,
					"errorType":    "ProviderUnavailable",
					"errorMessage": "provider is unavailable",
				},
			})
		default:
			http.NotFound(writer, request)
		}
	})

	client := newTestClient(
		t,
		"https://zarin-hub.test",
		handlerHTTPClient(handler),
		nil,
	)
	_, err := client.Convert(context.Background(), validClientRequest())
	require.Error(t, err)

	var clientError *cardclient.CardToIBANError
	require.ErrorAs(t, err, &clientError)
	assert.Equal(t, cardclient.CardToIBANErrorKindUnavailable, clientError.Kind)
	require.NotNil(t, clientError.HTTPStatus)
	assert.Equal(t, http.StatusServiceUnavailable, *clientError.HTTPStatus)
	require.NotNil(t, clientError.MetaCode)
	assert.Equal(t, 7, *clientError.MetaCode)
	require.NotNil(t, clientError.ErrorType)
	assert.Equal(t, "ProviderUnavailable", *clientError.ErrorType)
	assert.JSONEq(t, `{
		"data": null,
		"meta": {
			"isSuccess": false,
			"code": 7,
			"errorType": "ProviderUnavailable",
			"errorMessage": "provider is unavailable"
		}
	}`, string(clientError.RawResponse))
}

func TestCardToIBANClientClassifiesTimeout(t *testing.T) {
	t.Parallel()

	httpClient := &http.Client{Transport: roundTripperFunc(func(
		request *http.Request,
	) (*http.Response, error) {
		if request.URL.Path == authenticationPath {
			return jsonHTTPResponse(http.StatusOK, `{
				"data": {
					"accessToken": "access-token",
					"tokenType": "Bearer",
					"expiresTn": 3600,
					"refreshToken": "refresh-token"
				},
				"meta": {"isSuccess": true, "code": 0}
			}`), nil
		}

		return nil, context.DeadlineExceeded
	})}

	client := newTestClient(t, "https://zarin-hub.test", httpClient, nil)
	_, err := client.Convert(context.Background(), validClientRequest())
	require.Error(t, err)

	var clientError *cardclient.CardToIBANError
	require.ErrorAs(t, err, &clientError)
	assert.Equal(t, cardclient.CardToIBANErrorKindTimeout, clientError.Kind)
	assert.ErrorIs(t, clientError, context.DeadlineExceeded)
}

func TestNewCardToIBANClientDoesNotMutateProvidedHTTPClient(t *testing.T) {
	t.Parallel()

	httpClient := &http.Client{}
	client := newTestClient(t, "https://zarin-hub.test", httpClient, nil)

	assert.Zero(t, httpClient.Timeout)
	assert.Equal(t, 2*time.Second, client.httpClient.Timeout)
	assert.NotSame(t, httpClient, client.httpClient)
}

func newTestClient(
	t *testing.T,
	baseURL string,
	httpClient *http.Client,
	clock func() time.Time,
) *CardToIBANClient {
	t.Helper()

	client, err := NewCardToIBANClient(Config{
		BaseURL:        baseURL,
		Username:       "api-user",
		Password:       "api-password",
		RequestTimeout: 2 * time.Second,
		HTTPClient:     httpClient,
		Clock:          clock,
	})
	require.NoError(t, err)

	return client
}

func validClientRequest() cardclient.CardToIBANRequest {
	return cardclient.CardToIBANRequest{
		Card:         "6037997512345670",
		TrackID:      uuid.MustParse("b19e297d-0e04-49c4-8a9b-24b801e28d91"),
		HandlingType: model.CardToIBANRequestHandlingTypeStandard,
	}
}

func writeTokenResponse(t *testing.T, writer http.ResponseWriter) {
	t.Helper()
	writeJSON(t, writer, http.StatusOK, map[string]any{
		"data": map[string]any{
			"accessToken":  "access-token",
			"tokenType":    "Bearer",
			"expiresTn":    3600,
			"refreshToken": "refresh-token",
		},
		"meta": map[string]any{"isSuccess": true, "code": 0},
	})
}

func writeSuccessfulCardResponse(t *testing.T, writer http.ResponseWriter) {
	t.Helper()
	writeJSON(t, writer, http.StatusOK, map[string]any{
		"data": map[string]any{
			"bankName": "بانک ملی",
			"iban":     "IR820170000000123456789001",
		},
		"meta": map[string]any{"isSuccess": true, "code": 0},
	})
}

func writeJSON(t *testing.T, writer http.ResponseWriter, status int, body any) {
	t.Helper()
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	assert.NoError(t, json.NewEncoder(writer).Encode(body))
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (function roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func handlerHTTPClient(handler http.Handler) *http.Client {
	return &http.Client{Transport: roundTripperFunc(func(
		request *http.Request,
	) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		return recorder.Result(), nil
	})}
}

func jsonHTTPResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
