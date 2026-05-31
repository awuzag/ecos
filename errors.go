package ecos

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/samber/oops"
)

// HTTPError reports a non-2xx HTTP response.
type HTTPError struct {
	StatusCode int
	Status     string
	Body       string
}

func (err *HTTPError) Error() string {
	return fmt.Sprintf("ecos: http error: status=%d %s", err.StatusCode, err.Status)
}

// DecodeError reports a response decoding failure.
type DecodeError struct {
	Op  string
	Err error
}

func (err *DecodeError) Error() string {
	return fmt.Sprintf("ecos: decode %s: %v", err.Op, err.Err)
}

func (err *DecodeError) Unwrap() error {
	return err.Err
}

// APIError reports an ECOS business error response.
type APIError struct {
	Code    string
	Message string
}

func (err *APIError) Error() string {
	return fmt.Sprintf("ecos: api error: code=%s message=%s", err.Code, err.Message)
}

// RequestError reports a request failure with secrets redacted.
type RequestError struct {
	Op       string
	Endpoint string
	Err      string
}

func (err *RequestError) Error() string {
	if err.Endpoint != "" && err.Endpoint != err.Op {
		return fmt.Sprintf("ecos: request %s endpoint=%s: %s", err.Op, err.Endpoint, err.Err)
	}
	return fmt.Sprintf("ecos: request %s: %s", err.Op, err.Err)
}

func requestError(method string, endpoint string, op string, err error, apiKey string) error {
	if err == nil {
		return nil
	}
	safeEndpoint := redactSecret(endpoint, apiKey)
	message := err.Error()
	if strings.TrimSpace(apiKey) != "" {
		message = strings.ReplaceAll(message, apiKey, "[REDACTED]")
	}
	requestOp := method
	if strings.TrimSpace(requestOp) == "" {
		requestOp = op
	}
	return errorBuilder("request", method, safeEndpoint, op).
		Wrap(&RequestError{Op: requestOp, Endpoint: safeEndpoint, Err: message})
}

func checkHTTP(resp *resty.Response, method string, endpoint string, op string) error {
	if resp.IsSuccess() {
		return nil
	}
	return errorBuilder("http", method, endpoint, op).
		Wrap(&HTTPError{
			StatusCode: resp.StatusCode(),
			Status:     resp.Status(),
			Body:       string(resp.Body()),
		})
}

type errorEnvelope struct {
	Code    string `json:"CODE"`
	Message string `json:"MESSAGE"`
}

type ecosErrorResponse struct {
	Error errorEnvelope `json:"RESULT"`
}

func decodeBusinessError(body []byte, method string, endpoint string, op string) error {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" || !strings.HasPrefix(trimmed, "{") {
		return nil
	}

	var envelope ecosErrorResponse
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil
	}
	if envelope.Error.Code == "" {
		return nil
	}
	return errorBuilder("ecos_api", method, endpoint, op).
		With("code", envelope.Error.Code, "message", envelope.Error.Message).
		Wrap(&APIError{Code: envelope.Error.Code, Message: envelope.Error.Message})
}

func decodeError(method string, endpoint string, op string, format string, err error) error {
	if err == nil {
		return nil
	}
	return errorBuilder("decode", method, endpoint, op).
		With("format", format).
		Wrap(&DecodeError{Op: op, Err: err})
}

func errorBuilder(domain string, method string, endpoint string, op string) oops.OopsErrorBuilder {
	builder := oops.In(domain)
	attrs := make([]any, 0, 6)
	if method != "" {
		attrs = append(attrs, "method", method)
	}
	if endpoint != "" {
		attrs = append(attrs, "endpoint", endpoint)
	}
	if op != "" {
		attrs = append(attrs, "op", op)
	}
	if len(attrs) == 0 {
		return builder
	}
	return builder.With(attrs...)
}

func redactSecret(value string, secret string) string {
	if strings.TrimSpace(secret) == "" {
		return value
	}
	return strings.ReplaceAll(value, secret, "[REDACTED]")
}
