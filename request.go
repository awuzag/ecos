package ecos

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
)

const (
	defaultFormat = "json"
	defaultLang   = "kr"
)

func getJSON(ctx context.Context, client *Client, service string, pathSegments []string, out any) error {
	body, _, err := getRawJSON(ctx, client, service, pathSegments)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, out); err != nil {
		endpoint := buildEndpoint(service, client.apiKey, defaultFormat, defaultLang, pathSegments...)
		return decodeError("GET", redactSecret(endpoint, client.apiKey), service, defaultFormat, err)
	}
	return nil
}

func getRawJSON(ctx context.Context, client *Client, service string, pathSegments []string) ([]byte, string, error) {
	endpoint := buildEndpoint(service, client.apiKey, defaultFormat, defaultLang, pathSegments...)
	safeEndpoint := redactSecret(endpoint, client.apiKey)
	resp, err := client.resty.R().
		SetContext(ctx).
		Get(endpoint)
	if err != nil {
		return nil, safeEndpoint, requestError("GET", endpoint, service, err, client.apiKey)
	}
	if err := checkHTTP(resp, "GET", safeEndpoint, service); err != nil {
		return nil, safeEndpoint, err
	}

	body := resp.Body()
	if err := decodeBusinessError(body, "GET", safeEndpoint, service); err != nil {
		return nil, safeEndpoint, err
	}
	return body, safeEndpoint, nil
}

func buildEndpoint(service string, apiKey string, format string, lang string, pathSegments ...string) string {
	segments := make([]string, 0, 4+len(pathSegments))
	segments = append(segments, service, apiKey, format, lang)
	for _, segment := range pathSegments {
		if strings.TrimSpace(segment) != "" {
			segments = append(segments, segment)
		}
	}

	escaped := make([]string, 0, len(segments))
	for _, segment := range segments {
		escaped = append(escaped, url.PathEscape(segment))
	}
	return "/" + strings.Join(escaped, "/")
}
