package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/awuzag/ecos"
	"github.com/stretchr/testify/require"
)

func TestBuildSeriesSingleGroup(t *testing.T) {
	series := buildSeries([]catalogItem{
		{
			StatCode:  "722Y001",
			GroupCode: "Group1",
			ItemCode:  "0101000",
			Cycle:     "M",
			StartTime: "199905",
			EndTime:   "202605",
			UnitName:  "annual percent",
		},
	})

	require.Equal(t, []catalogSeries{
		{
			SeriesKey: "722Y001:M:0101000",
			StatCode:  "722Y001",
			Cycle:     "M",
			ItemCodes: []string{"0101000"},
			UnitName:  "annual percent",
			StartTime: "199905",
			EndTime:   "202605",
			Status:    "queryable",
		},
	}, series)
}

func TestBuildSeriesMultiGroupCandidate(t *testing.T) {
	series := buildSeries([]catalogItem{
		{
			StatCode:  "731Y006",
			GroupCode: "Group1",
			ItemCode:  "0000003",
			Cycle:     "M",
			StartTime: "199003",
			EndTime:   "202605",
			UnitName:  "KRW",
		},
		{
			StatCode:  "731Y006",
			GroupCode: "Group2",
			ItemCode:  "0000200",
			Cycle:     "M",
			StartTime: "199003",
			EndTime:   "202605",
			UnitName:  "KRW",
		},
	})

	require.Equal(t, []catalogSeries{
		{
			SeriesKey: "731Y006:M:0000003:0000200",
			StatCode:  "731Y006",
			Cycle:     "M",
			ItemCodes: []string{"0000003", "0000200"},
			UnitName:  "KRW",
			StartTime: "199003",
			EndTime:   "202605",
			Status:    "needs_review",
		},
	}, series)
}

func TestBuildSeriesDoesNotAutoExpandMultiGroupCartesianProduct(t *testing.T) {
	series := buildSeries([]catalogItem{
		{StatCode: "731Y006", GroupCode: "Group1", ItemCode: "0000003", Cycle: "M", StartTime: "199003"},
		{StatCode: "731Y006", GroupCode: "Group1", ItemCode: "0000004", Cycle: "M", StartTime: "199003"},
		{StatCode: "731Y006", GroupCode: "Group2", ItemCode: "0000200", Cycle: "M", StartTime: "199003"},
	})

	require.Empty(t, series)
}

func TestParseFlagsDefaultsToConservativeItemCollection(t *testing.T) {
	opts := parseFlags(nil)

	require.True(t, opts.includeKey)
	require.False(t, opts.allowFullItems)
	require.Equal(t, 1, opts.concurrency)
}

func TestValidateOptionsRequiresTargetForItemCollection(t *testing.T) {
	err := validateOptions(options{includeItems: true, concurrency: 1})

	require.ErrorContains(t, err, "include-items requires at least one -stat-code")
}

func TestValidateOptionsAllowsExplicitFullItemCrawl(t *testing.T) {
	err := validateOptions(options{includeItems: true, allowFullItems: true, concurrency: 1})

	require.NoError(t, err)
}

func TestNormalizedStatCodesSortsAndDeduplicates(t *testing.T) {
	require.Equal(t, []string{"722Y001", "817Y002"}, normalizedStatCodes([]string{"817Y002", " 722Y001 ", "817Y002"}))
}

func TestItemCollectionTargetsRespectsStartAndLimit(t *testing.T) {
	tables := []ecos.StatisticTable{
		{StatCode: "AAA", Searchable: ecos.SearchableYes},
		{StatCode: "SKIP", Searchable: ecos.SearchableNo},
		{StatCode: "BBB", Searchable: ecos.SearchableYes},
		{StatCode: "CCC", Searchable: ecos.SearchableYes},
	}

	targets, err := itemCollectionTargets(options{startStatCode: "BBB", maxItemTables: 2}, tables)

	require.NoError(t, err)
	require.Equal(t, []itemCollectionTarget{
		{Order: 0, StatCode: "BBB"},
		{Order: 1, StatCode: "CCC"},
	}, targets)
}

func TestItemCollectionTargetsCanLimitToExplicitStatCodes(t *testing.T) {
	tables := []ecos.StatisticTable{
		{StatCode: "AAA", Searchable: ecos.SearchableYes},
		{StatCode: "SKIP", Searchable: ecos.SearchableNo},
		{StatCode: "BBB", Searchable: ecos.SearchableYes},
		{StatCode: "CCC", Searchable: ecos.SearchableYes},
	}

	targets, err := itemCollectionTargets(options{targetStatCodes: stringList{"CCC", "AAA"}}, tables)

	require.NoError(t, err)
	require.Equal(t, []itemCollectionTarget{
		{Order: 0, StatCode: "AAA"},
		{Order: 1, StatCode: "CCC"},
	}, targets)
}

func TestApplyCollectedItemsOrdersByTargetOrder(t *testing.T) {
	result := inventory{Summary: inventorySummary{ItemTableTotal: 1}}
	baseItems := []catalogItem{{StatCode: "AAA", ItemCode: "1"}}
	baseSeries := []catalogSeries{{SeriesKey: "AAA:M:1", StatCode: "AAA"}}
	successes := map[int]itemCollectionResult{
		1: {
			Items:  []catalogItem{{StatCode: "CCC", ItemCode: "3"}},
			Series: []catalogSeries{{SeriesKey: "CCC:M:3", StatCode: "CCC"}},
		},
		0: {
			Items:  []catalogItem{{StatCode: "BBB", ItemCode: "2"}},
			Series: []catalogSeries{{SeriesKey: "BBB:M:2", StatCode: "BBB"}},
		},
	}

	applyCollectedItems(&result, baseItems, baseSeries, 1, successes)

	require.Equal(t, []catalogItem{
		{StatCode: "AAA", ItemCode: "1"},
		{StatCode: "BBB", ItemCode: "2"},
		{StatCode: "CCC", ItemCode: "3"},
	}, result.Items)
	require.Equal(t, 3, result.Summary.ItemTableTotal)
	require.Equal(t, 3, result.Summary.ItemTotal)
	require.Equal(t, 3, result.Summary.SeriesTotal)
}

func TestWriteInventoryFileReplacesSnapshot(t *testing.T) {
	path := filepath.Join(t.TempDir(), "inventory.json")

	first := inventory{
		Version: inventoryVersion,
		Source: inventorySource{
			Provider:    sourceProvider,
			CollectedAt: "2026-06-04T00:00:00Z",
			Format:      sourceFormat,
			Services:    []string{"StatisticTableList"},
		},
		Summary: inventorySummary{TableTotal: 834},
		Tables: []catalogTable{
			{StatCode: "722Y001", StatName: "base rate", Cycle: "M", Searchable: true},
		},
		Items:  []catalogItem{},
		Series: []catalogSeries{},
	}
	require.NoError(t, writeInventoryFile(path, first))

	second := first
	second.Source.Services = append(second.Source.Services, "StatisticItemList")
	second.Items = []catalogItem{
		{StatCode: "722Y001", GroupCode: "Group1", ItemCode: "0101000", Cycle: "M"},
	}
	second.Series = []catalogSeries{
		{SeriesKey: "722Y001:M:0101000", StatCode: "722Y001", Cycle: "M", ItemCodes: []string{"0101000"}, Status: "queryable"},
	}
	second.Summary.ItemTableTotal = 1
	second.Summary.ItemTotal = 1
	second.Summary.SeriesTotal = 1
	require.NoError(t, writeInventoryFile(path, second))

	body, err := os.ReadFile(path)
	require.NoError(t, err)

	var decoded inventory
	require.NoError(t, json.Unmarshal(body, &decoded))
	require.Equal(t, 1, decoded.Summary.ItemTableTotal)
	require.Equal(t, []catalogSeries{
		{SeriesKey: "722Y001:M:0101000", StatCode: "722Y001", Cycle: "M", ItemCodes: []string{"0101000"}, Status: "queryable"},
	}, decoded.Series)
}

func TestApplyResumeSnapshotKeepsPreviousTables(t *testing.T) {
	path := filepath.Join(t.TempDir(), "inventory.json")
	existing := inventory{
		Version: inventoryVersion,
		Source: inventorySource{
			Provider:    sourceProvider,
			CollectedAt: "2026-06-04T00:00:00Z",
			Format:      sourceFormat,
			Services:    []string{"StatisticTableList", "StatisticItemList"},
		},
		Summary: inventorySummary{TableTotal: 3, SearchableTableTotal: 3, ItemTableTotal: 2, ItemTotal: 2, SeriesTotal: 2},
		Tables: []catalogTable{
			{StatCode: "AAA", Searchable: true},
			{StatCode: "BBB", Searchable: true},
			{StatCode: "CCC", Searchable: true},
		},
		Items: []catalogItem{
			{StatCode: "AAA", ItemCode: "1"},
			{StatCode: "BBB", ItemCode: "2"},
		},
		Series: []catalogSeries{
			{SeriesKey: "AAA:M:1", StatCode: "AAA"},
			{SeriesKey: "BBB:M:2", StatCode: "BBB"},
		},
	}
	require.NoError(t, writeInventoryFile(path, existing))

	next := inventory{
		Version: inventoryVersion,
		Source:  existing.Source,
		Summary: inventorySummary{TableTotal: 3, SearchableTableTotal: 3},
		Tables:  existing.Tables,
		Items:   []catalogItem{},
		Series:  []catalogSeries{},
	}
	tables := []ecos.StatisticTable{
		{StatCode: "AAA", Searchable: ecos.SearchableYes},
		{StatCode: "BBB", Searchable: ecos.SearchableYes},
		{StatCode: "CCC", Searchable: ecos.SearchableYes},
	}

	require.NoError(t, applyResumeSnapshot(path, "BBB", tables, &next))

	require.Equal(t, 1, next.Summary.ItemTableTotal)
	require.Equal(t, []catalogItem{{StatCode: "AAA", ItemCode: "1"}}, next.Items)
	require.Equal(t, []catalogSeries{{SeriesKey: "AAA:M:1", StatCode: "AAA"}}, next.Series)
}
