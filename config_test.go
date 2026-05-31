package ecos

import (
	"testing"

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
