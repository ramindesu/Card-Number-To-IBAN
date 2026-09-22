package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/google/uuid"

	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/service"
)

const maximumRequestBodySize = 4 << 10

type cardToIBANService interface {
	Convert(context.Context, service.CardToIBANRequest) (service.CardToIBANResponse, error)
}

// CardToIBANHandler maps HTTP requests to the Card-to-IBAN application service.
type CardToIBANHandler struct {
	service cardToIBANService
}

// NewCardToIBANHandler creates the HTTP handler for this vertical slice.
func NewCardToIBANHandler(applicationService cardToIBANService) (*CardToIBANHandler, error) {
	if applicationService == nil {
		return nil, errors.New("card-to-IBAN service is required")
	}

	return &CardToIBANHandler{service: applicationService}, nil
}

// Convert handles a Card-to-IBAN request.
func (h *CardToIBANHandler) Convert(writer http.ResponseWriter, request *http.Request) {
	if !hasJSONContentType(request) {
		writeError(
			writer,
			http.StatusUnsupportedMediaType,
			"unsupported_media_type",
			"Content-Type باید application/json باشد",
			nil,
			"",
		)
		return
	}

	payload, err := decodeRequest(writer, request)
	if err != nil {
		var maximumBytesError *http.MaxBytesError
		if errors.As(err, &maximumBytesError) {
			writeError(
				writer,
				http.StatusRequestEntityTooLarge,
				"request_too_large",
				"حجم بدنه درخواست بیش از حد مجاز است",
				nil,
				"",
			)
			return
		}

		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_request",
			"بدنه درخواست JSON معتبری نیست",
			nil,
			"",
		)
		return
	}

	result, err := h.service.Convert(request.Context(), service.CardToIBANRequest{
		Card:     payload.Card,
		ClientIP: parseClientIP(request.RemoteAddr),
	})
	if err != nil {
		writeServiceError(writer, err)
		return
	}

	writer.Header().Set("X-Request-ID", result.RequestID.String())
	writeJSON(writer, http.StatusOK, convertResponse{
		Data: convertResponseData{
			RequestID:          result.RequestID,
			TrackID:            result.TrackID,
			BankName:           result.BankName,
			EnglishBankName:    result.EnglishBankName,
			IBAN:               result.IBAN,
			Deposit:            result.Deposit,
			DepositDescription: result.DepositDescription,
			DepositOwners:      result.DepositOwners,
			DepositStatus:      result.DepositStatus,
		},
	})
}

func hasJSONContentType(request *http.Request) bool {
	mediaType, _, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	return err == nil && mediaType == "application/json"
}

func decodeRequest(writer http.ResponseWriter, request *http.Request) (convertRequest, error) {
	request.Body = http.MaxBytesReader(writer, request.Body, maximumRequestBodySize)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()

	var payload convertRequest
	if err := decoder.Decode(&payload); err != nil {
		return convertRequest{}, err
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return convertRequest{}, errors.New("request body must contain exactly one JSON value")
	}

	return payload, nil
}

func writeServiceError(writer http.ResponseWriter, err error) {
	var serviceError *service.CardToIBANError
	if !errors.As(err, &serviceError) {
		writeError(
			writer,
			http.StatusInternalServerError,
			string(service.CardToIBANErrorCodeInternal),
			"خطای داخلی رخ داده است",
			nil,
			"",
		)
		return
	}

	requestID := serviceError.RequestID
	writer.Header().Set("X-Request-ID", requestID.String())
	writeError(
		writer,
		serviceError.HTTPStatus,
		string(serviceError.Code),
		serviceError.Message,
		&requestID,
		string(serviceError.ValidationCode),
	)
}

func writeError(
	writer http.ResponseWriter,
	status int,
	code string,
	message string,
	requestID *uuid.UUID,
	validationCode string,
) {
	writeJSON(writer, status, errorResponse{
		Error: errorResponseData{
			Code:           code,
			Message:        message,
			RequestID:      requestID,
			ValidationCode: validationCode,
		},
	})
}

func writeJSON(writer http.ResponseWriter, status int, payload any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(payload)
}
