package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/cardnumber"
	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/handler"
	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/service"
)

func TestConvertReturnsSuccessfulResponse(t *testing.T) {
	t.Parallel()

	requestID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	trackID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	applicationService := &fakeCardToIBANService{
		response: service.CardToIBANResponse{
			RequestID:          requestID,
			TrackID:            trackID,
			BankName:           "بانک ملی ایران",
			EnglishBankName:    "Bank Melli Iran",
			IBAN:               "IR820540102680020817909002",
			Deposit:            "0123456789",
			DepositDescription: "حساب کوتاه مدت",
			DepositOwners:      "کاربر تست",
			DepositStatus:      "active",
		},
	}
	router := newTestRouter(t, applicationService)
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/card-to-iban",
		strings.NewReader(`{"card":"۶۰۳۷-۹۹۷۵-۱۲۳۴-۵۶۷۰"}`),
	)
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.RemoteAddr = "192.0.2.10:42100"
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json; charset=utf-8", recorder.Header().Get("Content-Type"))
	assert.Equal(t, requestID.String(), recorder.Header().Get("X-Request-ID"))
	assert.JSONEq(t, `{
		"data": {
			"requestId": "11111111-1111-4111-8111-111111111111",
			"trackId": "22222222-2222-4222-8222-222222222222",
			"bankName": "بانک ملی ایران",
			"englishBankName": "Bank Melli Iran",
			"iban": "IR820540102680020817909002",
			"deposit": "0123456789",
			"depositDescription": "حساب کوتاه مدت",
			"depositOwners": "کاربر تست",
			"depositStatus": "active"
		}
	}`, recorder.Body.String())
	require.Len(t, applicationService.requests, 1)
	assert.Equal(t, "۶۰۳۷-۹۹۷۵-۱۲۳۴-۵۶۷۰", applicationService.requests[0].Card)
	require.NotNil(t, applicationService.requests[0].ClientIP)
	assert.Equal(t, netip.MustParseAddr("192.0.2.10"), *applicationService.requests[0].ClientIP)
}

func TestConvertMapsServiceValidationError(t *testing.T) {
	t.Parallel()

	requestID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	applicationService := &fakeCardToIBANService{
		err: &service.CardToIBANError{
			Code:           service.CardToIBANErrorCodeValidation,
			Message:        "شماره کارت باید ۱۶ رقم باشد",
			HTTPStatus:     http.StatusUnprocessableEntity,
			RequestID:      requestID,
			ValidationCode: cardnumber.ValidationErrorCodeInvalidLength,
		},
	}
	router := newTestRouter(t, applicationService)
	request := newJSONRequest(`{"card":"1234"}`)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
	assert.Equal(t, requestID.String(), recorder.Header().Get("X-Request-ID"))
	assert.JSONEq(t, `{
		"error": {
			"code": "validation_error",
			"message": "شماره کارت باید ۱۶ رقم باشد",
			"requestId": "11111111-1111-4111-8111-111111111111",
			"validationCode": "card_invalid_length"
		}
	}`, recorder.Body.String())
	assert.Len(t, applicationService.requests, 1)
}

func TestConvertRejectsMalformedRequestBeforeCallingService(t *testing.T) {
	t.Parallel()

	testCases := map[string]string{
		"malformed JSON": `{"card":`,
		"unknown field":  `{"card":"1234","password":"secret"}`,
		"multiple values": `{"card":"1234"}
			{"card":"5678"}`,
	}

	for name, body := range testCases {
		body := body
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			applicationService := &fakeCardToIBANService{}
			router := newTestRouter(t, applicationService)
			request := newJSONRequest(body)
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, request)

			assert.Equal(t, http.StatusBadRequest, recorder.Code)
			assert.JSONEq(t, `{
				"error": {
					"code": "invalid_request",
					"message": "بدنه درخواست JSON معتبری نیست"
				}
			}`, recorder.Body.String())
			assert.Empty(t, applicationService.requests)
		})
	}
}

func TestConvertRequiresJSONContentType(t *testing.T) {
	t.Parallel()

	applicationService := &fakeCardToIBANService{}
	router := newTestRouter(t, applicationService)
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/card-to-iban",
		strings.NewReader(`{"card":"1234"}`),
	)
	request.Header.Set("Content-Type", "text/plain")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusUnsupportedMediaType, recorder.Code)
	assert.JSONEq(t, `{
		"error": {
			"code": "unsupported_media_type",
			"message": "Content-Type باید application/json باشد"
		}
	}`, recorder.Body.String())
	assert.Empty(t, applicationService.requests)
}

func TestConvertRejectsOversizedRequest(t *testing.T) {
	t.Parallel()

	applicationService := &fakeCardToIBANService{}
	router := newTestRouter(t, applicationService)
	request := newJSONRequest(`{"card":"` + strings.Repeat("1", 5<<10) + `"}`)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusRequestEntityTooLarge, recorder.Code)
	assert.JSONEq(t, `{
		"error": {
			"code": "request_too_large",
			"message": "حجم بدنه درخواست بیش از حد مجاز است"
		}
	}`, recorder.Body.String())
	assert.Empty(t, applicationService.requests)
}

func TestConvertDoesNotExposeUnexpectedServiceError(t *testing.T) {
	t.Parallel()

	applicationService := &fakeCardToIBANService{
		err: errors.New("database password is secret"),
	}
	router := newTestRouter(t, applicationService)
	request := newJSONRequest(`{"card":"6037997512345670"}`)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.NotContains(t, recorder.Body.String(), "database password")
	assert.JSONEq(t, `{
		"error": {
			"code": "internal_error",
			"message": "خطای داخلی رخ داده است"
		}
	}`, recorder.Body.String())
}

func TestRegisterRoutesRejectsUnsupportedMethod(t *testing.T) {
	t.Parallel()

	router := newTestRouter(t, &fakeCardToIBANService{})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/card-to-iban", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusMethodNotAllowed, recorder.Code)
}

func newTestRouter(
	t *testing.T,
	applicationService *fakeCardToIBANService,
) http.Handler {
	t.Helper()

	httpHandler, err := handler.NewCardToIBANHandler(applicationService)
	require.NoError(t, err)

	router := chi.NewRouter()
	router.Route("/api/v1", httpHandler.RegisterRoutes)
	return router
}

func newJSONRequest(body string) *http.Request {
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/card-to-iban",
		strings.NewReader(body),
	)
	request.Header.Set("Content-Type", "application/json")
	return request
}

type fakeCardToIBANService struct {
	response service.CardToIBANResponse
	err      error
	requests []service.CardToIBANRequest
}

func (fake *fakeCardToIBANService) Convert(
	_ context.Context,
	request service.CardToIBANRequest,
) (service.CardToIBANResponse, error) {
	fake.requests = append(fake.requests, request)
	return fake.response, fake.err
}
