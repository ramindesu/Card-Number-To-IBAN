package zarinhub

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"

	cardclient "github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/client"
)

const (
	cardToIBANPath          = "/api/v5/KYC/CardToIban"
	defaultTokenRefreshSkew = 30 * time.Second
)

// Config contains the dependencies and settings of the ZarinHub client.
type Config struct {
	BaseURL          string
	Username         string
	Password         string
	RequestTimeout   time.Duration
	TokenRefreshSkew time.Duration
	HTTPClient       *http.Client
	Clock            func() time.Time
}

// CardToIBANClient implements the provider-independent client contract using
// ZarinHub API v5.
type CardToIBANClient struct {
	httpClient *http.Client
	baseURL    *url.URL
	tokens     *tokenManager
}

// NewCardToIBANClient creates a ZarinHub-backed Card-to-IBAN client.
func NewCardToIBANClient(config Config) (*CardToIBANClient, error) {
	baseURL, err := url.Parse(config.BaseURL)
	if err != nil ||
		(baseURL.Scheme != "http" && baseURL.Scheme != "https") ||
		baseURL.Host == "" {
		return nil, errors.New("ZarinHub base URL is invalid")
	}

	if config.Username == "" || config.Password == "" {
		return nil, errors.New("ZarinHub API credentials are required")
	}

	if config.RequestTimeout <= 0 {
		return nil, errors.New("ZarinHub request timeout must be positive")
	}

	if config.TokenRefreshSkew <= 0 {
		config.TokenRefreshSkew = defaultTokenRefreshSkew
	}

	if config.Clock == nil {
		config.Clock = func() time.Time { return time.Now().UTC() }
	}

	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{Timeout: config.RequestTimeout}
	} else if config.HTTPClient.Timeout == 0 {
		httpClient := *config.HTTPClient
		httpClient.Timeout = config.RequestTimeout
		config.HTTPClient = &httpClient
	}

	client := &CardToIBANClient{
		httpClient: config.HTTPClient,
		baseURL:    baseURL,
	}
	client.tokens = &tokenManager{
		httpClient:  config.HTTPClient,
		baseURL:     baseURL,
		username:    config.Username,
		password:    config.Password,
		refreshSkew: config.TokenRefreshSkew,
		clock:       config.Clock,
	}

	return client, nil
}

// Convert calls ZarinHub CardToIban with a cached Bearer token.
func (c *CardToIBANClient) Convert(
	ctx context.Context,
	request cardclient.CardToIBANRequest,
) (cardclient.CardToIBANResult, error) {
	authorization, err := c.tokens.authorization(ctx)
	if err != nil {
		return cardclient.CardToIBANResult{}, err
	}

	httpStatus, rawResponse, err := postJSON(
		ctx,
		c.httpClient,
		c.baseURL,
		cardToIBANPath,
		authorization,
		cardToIBANRequest{
			Card:                request.Card,
			TrackID:             request.TrackID,
			RequestHandlingType: string(request.HandlingType),
		},
	)
	if err != nil {
		return cardclient.CardToIBANResult{}, requestError(err)
	}

	var response responseEnvelope[*cardToIBANData]
	if err := json.Unmarshal(rawResponse, &response); err != nil {
		return cardclient.CardToIBANResult{}, malformedResponseError(
			httpStatus,
			rawResponse,
			err,
		)
	}

	if httpStatus < http.StatusOK ||
		httpStatus >= http.StatusMultipleChoices ||
		!response.Meta.IsSuccess {
		return cardclient.CardToIBANResult{}, responseError(
			httpStatus,
			response.Meta,
			rawResponse,
		)
	}

	if response.Data == nil || response.Data.IBAN == "" {
		return cardclient.CardToIBANResult{}, malformedResponseError(
			httpStatus,
			rawResponse,
			errors.New("successful response is missing IBAN data"),
		)
	}

	return cardclient.CardToIBANResult{
		HTTPStatus: httpStatus,
		MetaCode:   response.Meta.Code,
		Data: cardclient.CardToIBANData{
			BankName:           response.Data.BankName,
			EnglishBankName:    response.Data.EnglishBankName,
			IBAN:               response.Data.IBAN,
			Deposit:            response.Data.Deposit,
			DepositDescription: response.Data.DepositDescription,
			DepositOwners:      response.Data.DepositOwners,
			DepositStatus:      response.Data.DepositStatus,
		},
		RawResponse: append(json.RawMessage(nil), rawResponse...),
	}, nil
}

var _ cardclient.CardToIBANClient = (*CardToIBANClient)(nil)
