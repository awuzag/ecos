package ecos

import (
	"net/http"
	"strings"
	"time"

	"github.com/samber/oops"
)

const defaultBaseURL = "https://ecos.bok.or.kr/api"

// Config holds required ECOS client settings.
type Config struct {
	// APIKey is the 한국은행 ECOS OpenAPI authentication key.
	APIKey string
}

// Option customizes a Client.
type Option func(*clientOptions)

type clientOptions struct {
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
}

// WithBaseURL overrides the ECOS base URL.
func WithBaseURL(baseURL string) Option {
	return func(options *clientOptions) {
		options.baseURL = baseURL
	}
}

// WithHTTPClient uses a caller-provided HTTP client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(options *clientOptions) {
		options.httpClient = httpClient
	}
}

// WithTimeout sets resty's request timeout when no custom HTTP client is provided.
func WithTimeout(timeout time.Duration) Option {
	return func(options *clientOptions) {
		options.timeout = timeout
	}
}

func buildOptions(opts []Option) clientOptions {
	options := clientOptions{
		baseURL: defaultBaseURL,
		timeout: 30 * time.Second,
	}
	for _, opt := range opts {
		opt(&options)
	}
	options.baseURL = strings.TrimRight(options.baseURL, "/")
	return options
}

func validateConfig(config Config) error {
	if strings.TrimSpace(config.APIKey) == "" {
		return oops.In("config").
			With("field", "APIKey").
			New("ecos: APIKey is required")
	}
	return nil
}
