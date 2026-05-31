package ecos

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequestErrorRedactsAPIKey(t *testing.T) {
	err := requestError(
		"GET",
		"/StatisticSearch/super-secret/json/kr/1/10",
		"StatisticSearch",
		errors.New("Get /StatisticSearch/super-secret/json/kr/1/10: connection refused"),
		"super-secret",
	)

	require.Error(t, err)
	require.NotContains(t, err.Error(), "super-secret")
	require.Contains(t, err.Error(), "[REDACTED]")
}
