//go:build e2e

package ecos

import (
	"bufio"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestE2ESmoke(t *testing.T) {
	apiKey := envOrDotEnv("ECOS_API_KEY")
	if strings.TrimSpace(apiKey) == "" {
		t.Skip("ECOS_API_KEY is required for e2e")
	}

	client, err := New(Config{APIKey: apiKey})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	t.Cleanup(cancel)

	tables, err := client.Tables(ctx, TablesRequest{Page: NewPage(1, 1)})
	require.NoError(t, err)
	require.NotEmpty(t, tables.Rows)

	items, err := client.Items(ctx, ItemsRequest{Page: NewPage(1, 1), StatCode: StatCodeBankOfKoreaBaseRate})
	require.NoError(t, err)
	require.NotEmpty(t, items.Rows)

	observations, err := client.Search(ctx, SearchRequest{
		Page:      NewPage(1, 1),
		StatCode:  StatCodeBankOfKoreaBaseRate,
		Cycle:     CycleMonthly,
		StartTime: "202001",
		EndTime:   "202001",
		ItemCodes: []ItemCode{ItemCodeBankOfKoreaBaseRate},
	})
	require.NoError(t, err)
	require.NotEmpty(t, observations.Rows)

	keyStatistics, err := client.KeyStatistics(ctx, KeyStatisticsRequest{Page: NewPage(1, 1)})
	require.NoError(t, err)
	require.NotEmpty(t, keyStatistics.Rows)
	require.NotEmpty(t, keyStatistics.Rows[0].ReferenceTime)

	meta, err := client.Meta(ctx, MetaRequest{Page: NewPage(1, 1), DataName: "경제심리지수"})
	require.NoError(t, err)
	require.NotNil(t, meta.Rows)

	words, err := client.Words(ctx, WordsRequest{Page: NewPage(1, 1), Word: "소비자동향지수"})
	require.NoError(t, err)
	require.NotNil(t, words.Rows)
}

func envOrDotEnv(key string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value != "" {
		return value
	}

	file, err := os.Open(".env")
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	prefix := key + "="
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || !strings.HasPrefix(line, prefix) {
			continue
		}
		return strings.Trim(strings.TrimPrefix(line, prefix), `"'`)
	}
	return ""
}
