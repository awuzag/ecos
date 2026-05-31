package ecos

import (
	"net/http"

	"github.com/go-resty/resty/v2"
)

// Client calls 한국은행 ECOS OpenAPI endpoints.
type Client struct {
	apiKey string
	resty  *resty.Client
}

// New creates an ECOS client.
func New(config Config, opts ...Option) (*Client, error) {
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	options := buildOptions(opts)
	httpClient := resty.NewWithClient(&http.Client{})
	if options.httpClient != nil {
		httpClient = resty.NewWithClient(options.httpClient)
	}

	httpClient.
		SetBaseURL(options.baseURL).
		SetHeader("User-Agent", "github.com/awuzag/ecos").
		SetTimeout(options.timeout)

	return &Client{
		apiKey: config.APIKey,
		resty:  httpClient,
	}, nil
}
