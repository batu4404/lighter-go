package http

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"

	"github.com/bytedance/sonic"
)

// maxRetries is the number of additional attempts made for transient failures
// (HTTP 429 and 5xx) on top of the initial request.
const maxRetries = 3

func (c *client) parseResultStatus(respBody []byte) error {
	resultStatus := &ResultCode{}
	if err := sonic.Unmarshal(respBody, resultStatus); err != nil {
		return err
	}
	if resultStatus.Code != CodeOK {
		return errors.New(resultStatus.Message)
	}
	return nil
}

// retryBackoff returns the delay before the given attempt (0-indexed):
// 200ms, 400ms, 800ms, ... capped at 5s.
func retryBackoff(attempt int) time.Duration {
	d := 200 * time.Millisecond * (1 << attempt)
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}

// isRetryableStatus reports whether an HTTP status warrants a retry.
func isRetryableStatus(code int) bool {
	return code == http.StatusTooManyRequests || code >= 500
}

// doWithRetry builds and executes an HTTP request with the given context,
// retrying on transient failures (429/5xx) with exponential backoff. contentType
// and body are used for POST requests; pass "" / nil for GET. The returned
// response body has not been read; callers own closing it.
func (c *client) doWithRetry(ctx context.Context, method, urlStr, contentType, auth string, body []byte) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryBackoff(attempt - 1)):
			}
		}

		var reader io.Reader
		if body != nil {
			reader = bytes.NewReader(body)
		}
		req, err := http.NewRequestWithContext(ctx, method, urlStr, reader)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
		if auth != "" {
			req.Header.Set("Authorization", auth)
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			// Context cancellation/deadline is terminal, not retryable.
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			lastErr = &ConnectionError{Err: err}
			continue
		}
		if isRetryableStatus(resp.StatusCode) && attempt < maxRetries {
			resp.Body.Close() //nolint:errcheck // discarding a retryable response
			lastErr = NewAPIErrorWithStatus(int32(resp.StatusCode), "transient", resp.StatusCode)
			continue
		}
		return resp, nil
	}
	return nil, lastErr
}

func (c *client) getAndParseL2HTTPResponse(ctx context.Context, path string, params map[string]any, result interface{}) error {
	return c.getAndParseL2HTTPResponseWithAuth(ctx, path, params, "", result)
}

func (c *client) getAndParseL2HTTPResponseWithAuth(ctx context.Context, path string, params map[string]any, auth string, result interface{}) error {
	u, err := url.Parse(c.endpoint)
	if err != nil {
		return err
	}
	u.Path = u.Path + "/" + path

	q := u.Query()
	for k, v := range params {
		q.Set(k, fmt.Sprintf("%v", v))
	}
	u.RawQuery = q.Encode()

	resp, err := c.doWithRetry(ctx, http.MethodGet, u.String(), "", auth, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck // Response body close errors are non-actionable
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		// Try to parse the error response for structured error info
		resultStatus := &ResultCode{}
		if err := sonic.Unmarshal(body, resultStatus); err == nil && resultStatus.Code != 0 {
			return fmt.Errorf("lighter API error (code %d, http %d): %s", resultStatus.Code, resp.StatusCode, resultStatus.Message)
		}
		return fmt.Errorf("lighter API error (http %d): %s", resp.StatusCode, string(body))
	}
	if err = c.parseResultStatus(body); err != nil {
		return err
	}
	if err := sonic.Unmarshal(body, result); err != nil {
		return err
	}
	return nil
}

func (c *client) GetNextNonce(ctx context.Context, accountIndex int64, apiKeyIndex uint8) (int64, error) {
	result := &NextNonce{}
	err := c.getAndParseL2HTTPResponse(ctx, "api/v1/nextNonce", map[string]any{"account_index": accountIndex, "api_key_index": apiKeyIndex}, result)
	if err != nil {
		return -1, err
	}
	return result.Nonce, nil
}

func (c *client) GetApiKey(ctx context.Context, accountIndex int64, apiKeyIndex uint8) (string, error) {
	result := &AccountApiKeys{}
	err := c.getAndParseL2HTTPResponse(ctx, "api/v1/apikeys", map[string]any{"account_index": accountIndex, "api_key_index": apiKeyIndex}, result)
	if err != nil {
		return "", err
	}
	if len(result.ApiKeys) == 0 {
		return "", fmt.Errorf("no api keys returned")
	}
	return result.ApiKeys[0].PublicKey, nil
}

// InvalidateApiKeys is a no-op for this client because GetApiKey always
// performs a fresh HTTP request (no API-key caching). It exists to satisfy
// the MinimalHTTPClient interface and keep parity with the upstream signer.
func (c *client) InvalidateApiKeys(accountIndex int64) {}

// postAndParseL2HTTPResponse sends a POST request with JSON body and parses the response
func (c *client) postAndParseL2HTTPResponse(ctx context.Context, path string, body interface{}, result interface{}) error {
	u, err := url.Parse(c.endpoint)
	if err != nil {
		return err
	}
	u.Path = u.Path + "/" + path

	jsonBody, err := sonic.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	resp, err := c.doWithRetry(ctx, http.MethodPost, u.String(), "application/json", "", jsonBody)
	if err != nil {
		return err
	}
	return c.parsePostResponse(resp, result)
}

// postFormL2HTTPResponse sends a POST request with multipart/form-data body and parses the response
func (c *client) postFormL2HTTPResponse(ctx context.Context, path string, params map[string]any, result interface{}) error {
	return c.postFormL2HTTPResponseWithAuth(ctx, path, params, "", result)
}

// postFormL2HTTPResponseWithAuth sends a POST request with multipart/form-data body and auth header
func (c *client) postFormL2HTTPResponseWithAuth(ctx context.Context, path string, params map[string]any, auth string, result interface{}) error {
	u, err := url.Parse(c.endpoint)
	if err != nil {
		return err
	}
	u.Path = u.Path + "/" + path

	// Build multipart form data (matching Python SDK's multipart/form-data)
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for k, v := range params {
		if err := writer.WriteField(k, fmt.Sprintf("%v", v)); err != nil {
			return fmt.Errorf("failed to write form field %s: %w", k, err)
		}
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	resp, err := c.doWithRetry(ctx, http.MethodPost, u.String(), writer.FormDataContentType(), auth, body.Bytes())
	if err != nil {
		return err
	}
	return c.parsePostResponse(resp, result)
}

// postURLEncodedL2HTTPResponseWithAuth sends a POST request with application/x-www-form-urlencoded body and auth header
func (c *client) postURLEncodedL2HTTPResponseWithAuth(ctx context.Context, path string, params map[string]any, auth string, result interface{}) error {
	u, err := url.Parse(c.endpoint)
	if err != nil {
		return err
	}
	u.Path = u.Path + "/" + path

	// Build URL-encoded form data
	formData := url.Values{}
	for k, v := range params {
		// For arrays/slices, serialize as JSON
		switch val := v.(type) {
		case []uint8:
			// Convert []uint8 to []int to avoid base64 encoding
			intSlice := make([]int, len(val))
			for i, b := range val {
				intSlice[i] = int(b)
			}
			jsonBytes, err := sonic.Marshal(intSlice)
			if err != nil {
				return fmt.Errorf("failed to marshal %s: %w", k, err)
			}
			formData.Set(k, string(jsonBytes))
		case []string:
			jsonBytes, err := sonic.Marshal(val)
			if err != nil {
				return fmt.Errorf("failed to marshal %s: %w", k, err)
			}
			formData.Set(k, string(jsonBytes))
		default:
			formData.Set(k, fmt.Sprintf("%v", v))
		}
	}

	resp, err := c.doWithRetry(ctx, http.MethodPost, u.String(), "application/x-www-form-urlencoded;charset=UTF-8", auth, []byte(formData.Encode()))
	if err != nil {
		return err
	}
	return c.parsePostResponse(resp, result)
}

// parsePostResponse reads and parses a POST response body, applying the shared
// status-code and result-status checks.
func (c *client) parsePostResponse(resp *http.Response, result interface{}) error {
	defer resp.Body.Close() //nolint:errcheck // Response body close errors are non-actionable

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return NewAPIErrorWithStatus(int32(resp.StatusCode), string(respBody), resp.StatusCode)
	}

	if err = c.parseResultStatus(respBody); err != nil {
		return err
	}

	if err := sonic.Unmarshal(respBody, result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return nil
}
