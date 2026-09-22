package zarinhub

import (
	"encoding/json"
	"net/http"

	cardclient "github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/client"
)

type cardClientErrorKind = cardclient.CardToIBANErrorKind

const (
	cardClientErrorKindRejected       = cardclient.CardToIBANErrorKindRejected
	cardClientErrorKindAuthentication = cardclient.CardToIBANErrorKindAuthentication
	cardClientErrorKindTimeout        = cardclient.CardToIBANErrorKindTimeout
	cardClientErrorKindUnavailable    = cardclient.CardToIBANErrorKindUnavailable
	cardClientErrorKindTransport      = cardclient.CardToIBANErrorKindTransport
)

func responseError(
	httpStatus int,
	meta responseMeta,
	rawResponse json.RawMessage,
) *cardclient.CardToIBANError {
	errorType := meta.ErrorType
	metaCode := meta.Code

	return &cardclient.CardToIBANError{
		Kind:        responseErrorKind(httpStatus, meta.Code),
		HTTPStatus:  &httpStatus,
		MetaCode:    &metaCode,
		ErrorType:   errorType,
		RawResponse: rawResponse,
	}
}

func malformedResponseError(
	httpStatus int,
	rawResponse json.RawMessage,
	cause error,
) *cardclient.CardToIBANError {
	return &cardclient.CardToIBANError{
		Kind:        responseErrorKind(httpStatus, 0),
		HTTPStatus:  &httpStatus,
		RawResponse: rawResponse,
		Cause:       cause,
	}
}

func requestError(cause error) *cardclient.CardToIBANError {
	return &cardclient.CardToIBANError{
		Kind:  transportErrorKind(cause),
		Cause: cause,
	}
}

func responseErrorKind(httpStatus, metaCode int) cardclient.CardToIBANErrorKind {
	switch {
	case httpStatus == http.StatusUnauthorized, metaCode == 6:
		return cardClientErrorKindAuthentication
	case metaCode == 14:
		return cardClientErrorKindTimeout
	case httpStatus == http.StatusBadGateway,
		httpStatus == http.StatusServiceUnavailable,
		httpStatus >= http.StatusInternalServerError,
		metaCode == 1,
		metaCode == 7,
		metaCode == 15:
		return cardClientErrorKindUnavailable
	default:
		return cardClientErrorKindRejected
	}
}
