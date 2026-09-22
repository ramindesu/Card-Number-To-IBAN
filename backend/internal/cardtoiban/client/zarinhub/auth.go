package zarinhub

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	cardclient "github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/client"
)

const authenticationPath = "/api/v5/Authentication/GetToken"

type cachedToken struct {
	accessToken  string
	tokenType    string
	refreshToken string
	refreshAt    time.Time
}

type tokenManager struct {
	mutex       sync.Mutex
	httpClient  *http.Client
	baseURL     *url.URL
	username    string
	password    string
	refreshSkew time.Duration
	clock       func() time.Time
	token       cachedToken
}

func (m *tokenManager) authorization(ctx context.Context) (string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.token.accessToken != "" && m.clock().Before(m.token.refreshAt) {
		return m.token.tokenType + " " + m.token.accessToken, nil
	}

	var (
		newToken cachedToken
		err      error
	)

	if m.token.refreshToken != "" {
		newToken, err = m.requestToken(ctx, tokenRequest{
			RefreshToken: m.token.refreshToken,
		})
		if err != nil && !isAuthenticationError(err) {
			return "", err
		}
	}

	if newToken.accessToken == "" {
		newToken, err = m.requestToken(ctx, tokenRequest{
			Username: m.username,
			Password: m.password,
		})
		if err != nil {
			return "", err
		}
	}

	m.token = newToken
	return m.token.tokenType + " " + m.token.accessToken, nil
}

func (m *tokenManager) requestToken(
	ctx context.Context,
	payload tokenRequest,
) (cachedToken, error) {
	httpStatus, rawResponse, err := postJSON(
		ctx,
		m.httpClient,
		m.baseURL,
		authenticationPath,
		"",
		payload,
	)
	if err != nil {
		return cachedToken{}, requestError(err)
	}

	var response responseEnvelope[tokenData]
	if err := json.Unmarshal(rawResponse, &response); err != nil {
		return cachedToken{}, malformedResponseError(httpStatus, rawResponse, err)
	}

	if httpStatus < http.StatusOK ||
		httpStatus >= http.StatusMultipleChoices ||
		!response.Meta.IsSuccess {
		return cachedToken{}, responseError(httpStatus, response.Meta, rawResponse)
	}

	if response.Data.AccessToken == "" || response.Data.ExpiresIn <= 0 {
		return cachedToken{}, requestError(errors.New("authentication response is missing token data"))
	}

	tokenType := strings.TrimSpace(response.Data.TokenType)
	if tokenType == "" {
		tokenType = "Bearer"
	}

	lifetime := time.Duration(response.Data.ExpiresIn) * time.Second
	refreshSkew := m.refreshSkew
	if refreshSkew >= lifetime {
		refreshSkew = lifetime / 2
	}

	return cachedToken{
		accessToken:  response.Data.AccessToken,
		tokenType:    tokenType,
		refreshToken: response.Data.RefreshToken,
		refreshAt:    m.clock().Add(lifetime - refreshSkew),
	}, nil
}

func isAuthenticationError(err error) bool {
	var clientError *cardclient.CardToIBANError
	return errors.As(err, &clientError) &&
		clientError.Kind == cardclient.CardToIBANErrorKindAuthentication
}
