package ecos

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildEndpointEscapesSegments(t *testing.T) {
	endpoint := buildEndpoint("StatisticSearch", "secret/key", "json", "kr", "1", "10", "200Y101", "A", "2024", "2025", "101 01")

	require.Equal(t, "/StatisticSearch/secret%2Fkey/json/kr/1/10/200Y101/A/2024/2025/101%2001", endpoint)
}

func TestGetJSONCallsECOSPath(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"StatisticSearch":{"list":[]}}`))
	}))
	t.Cleanup(server.Close)

	client, err := New(Config{APIKey: "sample"}, WithBaseURL(server.URL))
	require.NoError(t, err)

	var out map[string]any
	err = getJSON(context.Background(), client, "StatisticSearch", []string{"1", "10", "200Y101", "A", "2024", "2025"}, &out)

	require.NoError(t, err)
	require.Equal(t, "/StatisticSearch/sample/json/kr/1/10/200Y101/A/2024/2025", gotPath)
	require.Contains(t, out, "StatisticSearch")
}

func TestGetJSONReturnsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"RESULT":{"CODE":"INFO-100","MESSAGE":"invalid request"}}`))
	}))
	t.Cleanup(server.Close)

	client, err := New(Config{APIKey: "sample"}, WithBaseURL(server.URL))
	require.NoError(t, err)

	var out map[string]any
	err = getJSON(context.Background(), client, "StatisticSearch", []string{"1", "10"}, &out)

	require.Error(t, err)
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	require.Equal(t, "INFO-100", apiErr.Code)
	require.Equal(t, "invalid request", apiErr.Message)
}
