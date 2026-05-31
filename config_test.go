package ecos

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewRequiresAPIKey(t *testing.T) {
	client, err := New(Config{})

	require.Nil(t, client)
	require.Error(t, err)
	require.Contains(t, err.Error(), "APIKey")
}

func TestNewTrimsBaseURLTrailingSlash(t *testing.T) {
	client, err := New(Config{APIKey: "secret"}, WithBaseURL("https://example.test/api/"))

	require.NoError(t, err)
	require.Equal(t, "https://example.test/api", client.resty.BaseURL)
}

func TestNewAppliesHTTPClient(t *testing.T) {
	httpClient := &http.Client{Timeout: 2 * time.Second}

	client, err := New(Config{APIKey: "secret"}, WithHTTPClient(httpClient))

	require.NoError(t, err)
	require.Same(t, httpClient, client.resty.GetClient())
}

func TestNewAppliesTimeout(t *testing.T) {
	client, err := New(Config{APIKey: "secret"}, WithTimeout(3*time.Second))

	require.NoError(t, err)
	require.Equal(t, 3*time.Second, client.resty.GetClient().Timeout)
}
