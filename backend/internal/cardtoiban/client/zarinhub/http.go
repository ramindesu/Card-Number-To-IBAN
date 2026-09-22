package zarinhub

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
)

const maximumResponseBodySize = 1 << 20

func postJSON(
	ctx context.Context,
	httpClient *http.Client,
	baseURL *url.URL,
	path string,
	authorization string,
	payload any,
) (int, []byte, error) {
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return 0, nil, fmt.Errorf("marshal request body: %w", err)
	}

	endpoint := baseURL.ResolveReference(&url.URL{Path: path})
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint.String(),
		bytes.NewReader(requestBody),
	)
	if err != nil {
		return 0, nil, fmt.Errorf("create request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	if authorization != "" {
		request.Header.Set("Authorization", authorization)
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return 0, nil, err
	}
	defer response.Body.Close()

	limitedBody := io.LimitReader(response.Body, maximumResponseBodySize+1)
	responseBody, err := io.ReadAll(limitedBody)
	if err != nil {
		return response.StatusCode, nil, fmt.Errorf("read response body: %w", err)
	}

	if len(responseBody) > maximumResponseBodySize {
		return response.StatusCode, nil, errors.New("response body exceeds size limit")
	}

	return response.StatusCode, responseBody, nil
}

func transportErrorKind(err error) cardClientErrorKind {
	if errors.Is(err, context.DeadlineExceeded) {
		return cardClientErrorKindTimeout
	}

	var networkError net.Error
	if errors.As(err, &networkError) && networkError.Timeout() {
		return cardClientErrorKindTimeout
	}

	return cardClientErrorKindTransport
}
