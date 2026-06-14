package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/awuzag/ecos"
	"github.com/samber/oops"
)

const (
	inventoryVersion = 1
	sourceProvider   = "BOK ECOS"
	sourceFormat     = "json"

	inventoryScopeBase                  = "tables_and_key_statistics"
	inventoryScopeTargetedItems         = "targeted_items"
	inventoryScopeExperimentalFullItems = "experimental_full_items"
)

var groupCodePattern = regexp.MustCompile(`(?i)^group([1-4])$`)

type options struct {
	apiKey          string
	envFile         string
	baseURL         string
	outPath         string
	pageSize        int
	includeKey      bool
	includeItems    bool
	targetStatCodes stringList
	maxItemTables   int
	startStatCode   string
	allowFullItems  bool
	concurrency     int
	sleep           time.Duration
}

type inventory struct {
	Version       int                   `json:"version"`
	Source        inventorySource       `json:"source"`
	Summary       inventorySummary      `json:"summary"`
	KeyStatistics []catalogKeyStatistic `json:"key_statistics"`
	Tables        []catalogTable        `json:"tables"`
	Items         []catalogItem         `json:"items"`
	Series        []catalogSeries       `json:"series"`
}

type inventorySource struct {
	Provider        string   `json:"provider"`
	CollectedAt     string   `json:"collected_at"`
	Format          string   `json:"format"`
	Services        []string `json:"services"`
	Scope           string   `json:"scope"`
	TargetStatCodes []string `json:"target_stat_codes,omitempty"`
}

type inventorySummary struct {
	TableTotal           int `json:"table_total"`
	KeyStatisticTotal    int `json:"key_statistic_total"`
	SearchableTableTotal int `json:"searchable_table_total"`
	ItemTableTotal       int `json:"item_table_total"`
	ItemTotal            int `json:"item_total"`
	SeriesTotal          int `json:"series_total"`
}

type catalogKeyStatistic struct {
	ClassName     string `json:"class_name"`
	Name          string `json:"name"`
	Value         string `json:"value"`
	ReferenceTime string `json:"reference_time"`
	UnitName      string `json:"unit_name"`
}

type catalogTable struct {
	StatCode       string `json:"stat_code"`
	ParentStatCode string `json:"parent_stat_code"`
	StatName       string `json:"stat_name"`
	Cycle          string `json:"cycle"`
	Searchable     bool   `json:"searchable"`
	Organization   string `json:"organization"`
}

type catalogItem struct {
	StatCode       string `json:"stat_code"`
	StatName       string `json:"stat_name"`
	GroupCode      string `json:"grp_code"`
	GroupName      string `json:"grp_name"`
	ItemCode       string `json:"item_code"`
	ItemName       string `json:"item_name"`
	ParentItemCode string `json:"parent_item_code"`
	ParentItemName string `json:"parent_item_name"`
	Cycle          string `json:"cycle"`
	StartTime      string `json:"start_time"`
	EndTime        string `json:"end_time"`
	DataCount      int    `json:"data_count"`
	UnitName       string `json:"unit_name"`
	Weight         string `json:"weight"`
}

type catalogSeries struct {
	SeriesKey string   `json:"series_key"`
	StatCode  string   `json:"stat_code"`
	Cycle     string   `json:"cycle"`
	ItemCodes []string `json:"item_codes"`
	UnitName  string   `json:"unit_name"`
	StartTime string   `json:"start_time"`
	EndTime   string   `json:"end_time"`
	Status    string   `json:"status"`
}

type itemCollectionTarget struct {
	Order    int
	StatCode ecos.StatCode
}

type itemCollectionResult struct {
	Order    int
	StatCode ecos.StatCode
	Total    int
	Items    []catalogItem
	Series   []catalogSeries
	Err      error
}

type stringList []string

func (values *stringList) String() string {
	return strings.Join(*values, ",")
}

func (values *stringList) Set(value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return oops.In("collect_series_catalog_inventory").
			With("field", "stat_code").
			New("ecos: stat-code must not be empty")
	}
	*values = append(*values, trimmed)
	return nil
}

func main() {
	code := 0
	if err := run(context.Background(), os.Args[1:], os.Stdout, os.Stderr); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		code = 1
	}
	os.Exit(code)
}

func run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) error {
	opts := parseFlags(args)
	if err := validateOptions(opts); err != nil {
		return err
	}
	apiKey, err := resolveAPIKey(opts.apiKey, opts.envFile)
	if err != nil {
		return err
	}
	if strings.TrimSpace(apiKey) == "" {
		return oops.In("collect_series_catalog_inventory").
			With("field", "api_key").
			New("ecos: API key is required; set --api-key, ECOS_API_KEY, or repo-local .env")
	}

	clientOptions := make([]ecos.Option, 0, 1)
	if strings.TrimSpace(opts.baseURL) != "" {
		clientOptions = append(clientOptions, ecos.WithBaseURL(opts.baseURL))
	}
	client, err := ecos.New(ecos.Config{APIKey: apiKey}, clientOptions...)
	if err != nil {
		return oops.In("collect_series_catalog_inventory").Wrap(err)
	}

	tables, total, err := collectTables(ctx, client, opts.pageSize)
	if err != nil {
		return err
	}

	result := inventory{
		Version: inventoryVersion,
		Source: inventorySource{
			Provider:    sourceProvider,
			CollectedAt: time.Now().UTC().Format(time.RFC3339),
			Format:      sourceFormat,
			Services:    []string{"StatisticTableList"},
			Scope:       inventoryScopeBase,
		},
		Summary: inventorySummary{
			TableTotal: total,
		},
		KeyStatistics: []catalogKeyStatistic{},
		Tables:        make([]catalogTable, 0, len(tables)),
		Items:         []catalogItem{},
		Series:        []catalogSeries{},
	}
	for _, table := range tables {
		if table.Searchable == ecos.SearchableYes {
			result.Summary.SearchableTableTotal++
		}
		result.Tables = append(result.Tables, toCatalogTable(table))
	}

	if opts.includeKey {
		keyStatistics, keyTotal, err := collectKeyStatistics(ctx, client, opts.pageSize)
		if err != nil {
			return err
		}
		result.Source.Services = append(result.Source.Services, "KeyStatisticList")
		result.KeyStatistics = make([]catalogKeyStatistic, 0, len(keyStatistics))
		for _, row := range keyStatistics {
			result.KeyStatistics = append(result.KeyStatistics, toCatalogKeyStatistic(row))
		}
		result.Summary.KeyStatisticTotal = keyTotal
	}

	if opts.includeItems {
		result.Source.Services = append(result.Source.Services, "StatisticItemList")
		result.Source.Scope = inventoryScopeExperimentalFullItems
		result.Source.TargetStatCodes = normalizedStatCodes(opts.targetStatCodes)
		if len(result.Source.TargetStatCodes) > 0 {
			result.Source.Scope = inventoryScopeTargetedItems
		}
		if err := applyResumeSnapshot(opts.outPath, opts.startStatCode, tables, &result); err != nil {
			return err
		}
		if err := persistPartialSnapshot(opts.outPath, result); err != nil {
			return err
		}
		if err := collectCatalogItems(ctx, client, opts, tables, &result, stderr); err != nil {
			return err
		}
	}

	return writeInventory(opts.outPath, stdout, result)
}

func parseFlags(args []string) options {
	var opts options
	flags := flag.NewFlagSet("collect-series-catalog-inventory", flag.ExitOnError)
	flags.StringVar(&opts.apiKey, "api-key", "", "ECOS API key")
	flags.StringVar(&opts.envFile, "env-file", ".env", "env file path for ECOS_API_KEY")
	flags.StringVar(&opts.baseURL, "base-url", "", "ECOS API base URL")
	flags.StringVar(&opts.outPath, "out", "-", "output JSON path, or - for stdout")
	flags.IntVar(&opts.pageSize, "page-size", 1000, "ECOS page size")
	flags.BoolVar(&opts.includeKey, "include-key-statistics", true, "collect KeyStatisticList rows for the 100 major indicators")
	flags.BoolVar(&opts.includeItems, "include-items", false, "collect StatisticItemList rows for targeted searchable tables")
	flags.Var(&opts.targetStatCodes, "stat-code", "target StatisticItemList table code; repeatable")
	flags.IntVar(&opts.maxItemTables, "max-item-tables", 0, "limit item collection to the first N targeted/searchable tables; 0 means no limit")
	flags.StringVar(&opts.startStatCode, "start-stat-code", "", "resume targeted item collection at this searchable statistic table code")
	flags.BoolVar(&opts.allowFullItems, "allow-full-item-crawl", false, "allow experimental full StatisticItemList crawl when -include-items has no -stat-code")
	flags.IntVar(&opts.concurrency, "concurrency", 1, "number of concurrent StatisticItemList table collectors")
	flags.DurationVar(&opts.sleep, "sleep", 250*time.Millisecond, "sleep duration between item table requests")
	_ = flags.Parse(args)
	return opts
}

func validateOptions(opts options) error {
	if opts.concurrency < 1 {
		return oops.In("collect_series_catalog_inventory").
			With("field", "concurrency").
			New("ecos: concurrency must be greater than zero")
	}
	if opts.includeItems && len(normalizedStatCodes(opts.targetStatCodes)) == 0 && !opts.allowFullItems {
		return oops.In("collect_series_catalog_inventory").
			With("field", "stat_code").
			New("ecos: include-items requires at least one -stat-code; pass -allow-full-item-crawl for experimental full crawl")
	}
	return nil
}

func normalizedStatCodes(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		normalized = append(normalized, trimmed)
	}
	sort.Strings(normalized)
	return normalized
}

func collectTables(ctx context.Context, client *ecos.Client, pageSize int) ([]ecos.StatisticTable, int, error) {
	if pageSize < 1 {
		return nil, 0, oops.In("collect_series_catalog_inventory").
			With("field", "page_size").
			New("ecos: page-size must be greater than zero")
	}

	first, err := client.Tables(ctx, ecos.TablesRequest{Page: ecos.NewPage(1, pageSize)})
	if err != nil {
		return nil, 0, oops.In("collect_series_catalog_inventory").
			With("service", "StatisticTableList").
			Wrap(err)
	}

	tables := append([]ecos.StatisticTable{}, first.Rows...)
	for start := pageSize + 1; start <= first.TotalCount; start += pageSize {
		end := start + pageSize - 1
		if end > first.TotalCount {
			end = first.TotalCount
		}
		page, err := client.Tables(ctx, ecos.TablesRequest{Page: ecos.NewPage(start, end)})
		if err != nil {
			return nil, 0, oops.In("collect_series_catalog_inventory").
				With("service", "StatisticTableList", "start", start, "end", end).
				Wrap(err)
		}
		tables = append(tables, page.Rows...)
	}
	return tables, first.TotalCount, nil
}

func collectKeyStatistics(ctx context.Context, client *ecos.Client, pageSize int) ([]ecos.KeyStatistic, int, error) {
	if pageSize < 1 {
		return nil, 0, oops.In("collect_series_catalog_inventory").
			With("field", "page_size").
			New("ecos: page-size must be greater than zero")
	}

	first, err := client.KeyStatistics(ctx, ecos.KeyStatisticsRequest{Page: ecos.NewPage(1, pageSize)})
	if err != nil {
		return nil, 0, oops.In("collect_series_catalog_inventory").
			With("service", "KeyStatisticList").
			Wrap(err)
	}

	rows := append([]ecos.KeyStatistic{}, first.Rows...)
	for start := pageSize + 1; start <= first.TotalCount; start += pageSize {
		end := start + pageSize - 1
		if end > first.TotalCount {
			end = first.TotalCount
		}
		page, err := client.KeyStatistics(ctx, ecos.KeyStatisticsRequest{Page: ecos.NewPage(start, end)})
		if err != nil {
			return nil, 0, oops.In("collect_series_catalog_inventory").
				With("service", "KeyStatisticList", "start", start, "end", end).
				Wrap(err)
		}
		rows = append(rows, page.Rows...)
	}
	return rows, first.TotalCount, nil
}

func collectCatalogItems(
	ctx context.Context,
	client *ecos.Client,
	opts options,
	tables []ecos.StatisticTable,
	result *inventory,
	stderr io.Writer,
) error {
	targets, err := itemCollectionTargets(opts, tables)
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		return nil
	}

	existingTables := result.Summary.ItemTableTotal
	baseItems := append([]catalogItem{}, result.Items...)
	baseSeries := append([]catalogSeries{}, result.Series...)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	results := collectItemsConcurrently(ctx, client, opts, targets)
	successes := map[int]itemCollectionResult{}
	var firstErr error
	for collected := range results {
		if collected.Err != nil {
			if firstErr == nil {
				firstErr = collected.Err
				cancel()
			}
			continue
		}

		successes[collected.Order] = collected
		applyCollectedItems(result, baseItems, baseSeries, existingTables, successes)
		_, _ = fmt.Fprintf(stderr, "collected items stat_code=%s total=%d\n", collected.StatCode, collected.Total)
		if err := persistPartialSnapshot(opts.outPath, *result); err != nil {
			return err
		}
	}
	if firstErr != nil {
		return firstErr
	}
	return nil
}

func itemCollectionTargets(opts options, tables []ecos.StatisticTable) ([]itemCollectionTarget, error) {
	targets := []itemCollectionTarget{}
	targetStatCodes := statCodeFilter(opts.targetStatCodes)
	started := strings.TrimSpace(opts.startStatCode) == ""
	for _, table := range tables {
		if table.Searchable != ecos.SearchableYes {
			continue
		}
		if len(targetStatCodes) > 0 && !targetStatCodes[string(table.StatCode)] {
			continue
		}
		if !started {
			if string(table.StatCode) != strings.TrimSpace(opts.startStatCode) {
				continue
			}
			started = true
		}
		if opts.maxItemTables > 0 && len(targets) >= opts.maxItemTables {
			break
		}
		targets = append(targets, itemCollectionTarget{
			Order:    len(targets),
			StatCode: table.StatCode,
		})
	}
	if !started {
		return nil, oops.In("collect_series_catalog_inventory").
			With("stat_code", strings.TrimSpace(opts.startStatCode)).
			New("ecos: start-stat-code was not found among searchable tables")
	}
	if len(targetStatCodes) > 0 {
		if missing := missingTargetStatCodes(opts.targetStatCodes, targets); len(missing) > 0 {
			return nil, oops.In("collect_series_catalog_inventory").
				With("stat_code", strings.Join(missing, ",")).
				New("ecos: stat-code was not found among targeted searchable tables")
		}
	}
	return targets, nil
}

func statCodeFilter(values []string) map[string]bool {
	filter := map[string]bool{}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			filter[trimmed] = true
		}
	}
	return filter
}

func missingTargetStatCodes(values []string, targets []itemCollectionTarget) []string {
	found := map[string]bool{}
	for _, target := range targets {
		found[string(target.StatCode)] = true
	}
	missing := []string{}
	seen := map[string]bool{}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		if !found[trimmed] {
			missing = append(missing, trimmed)
		}
	}
	return missing
}

func collectItemsConcurrently(
	ctx context.Context,
	client *ecos.Client,
	opts options,
	targets []itemCollectionTarget,
) <-chan itemCollectionResult {
	workerCount := opts.concurrency
	if workerCount > len(targets) {
		workerCount = len(targets)
	}

	jobs := make(chan itemCollectionTarget)
	results := make(chan itemCollectionResult)
	var workers sync.WaitGroup
	workers.Add(workerCount)
	for range workerCount {
		go func() {
			defer workers.Done()
			for target := range jobs {
				rows, total, err := collectItemsForTable(ctx, client, target.StatCode, opts.pageSize)
				if err != nil {
					results <- itemCollectionResult{Order: target.Order, StatCode: target.StatCode, Err: err}
					continue
				}
				items := make([]catalogItem, 0, len(rows))
				for _, row := range rows {
					items = append(items, toCatalogItem(row))
				}
				results <- itemCollectionResult{
					Order:    target.Order,
					StatCode: target.StatCode,
					Total:    total,
					Items:    items,
					Series:   buildSeries(items),
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for index, target := range targets {
			if opts.sleep > 0 && index > 0 {
				timer := time.NewTimer(opts.sleep)
				select {
				case <-ctx.Done():
					timer.Stop()
					return
				case <-timer.C:
				}
			}
			select {
			case <-ctx.Done():
				return
			case jobs <- target:
			}
		}
	}()

	go func() {
		workers.Wait()
		close(results)
	}()

	return results
}

func applyCollectedItems(
	result *inventory,
	baseItems []catalogItem,
	baseSeries []catalogSeries,
	existingTables int,
	successes map[int]itemCollectionResult,
) {
	orders := make([]int, 0, len(successes))
	for order := range successes {
		orders = append(orders, order)
	}
	sort.Ints(orders)

	result.Items = append([]catalogItem{}, baseItems...)
	result.Series = append([]catalogSeries{}, baseSeries...)
	for _, order := range orders {
		result.Items = append(result.Items, successes[order].Items...)
		result.Series = append(result.Series, successes[order].Series...)
	}
	result.Summary.ItemTableTotal = existingTables + len(successes)
	result.Summary.ItemTotal = len(result.Items)
	result.Summary.SeriesTotal = len(result.Series)
}

func collectItemsForTable(ctx context.Context, client *ecos.Client, statCode ecos.StatCode, pageSize int) ([]ecos.StatisticItem, int, error) {
	first, err := client.Items(ctx, ecos.ItemsRequest{
		Page:     ecos.NewPage(1, pageSize),
		StatCode: statCode,
	})
	if err != nil {
		return nil, 0, oops.In("collect_series_catalog_inventory").
			With("service", "StatisticItemList", "stat_code", string(statCode)).
			Wrap(err)
	}

	items := append([]ecos.StatisticItem{}, first.Rows...)
	for start := pageSize + 1; start <= first.TotalCount; start += pageSize {
		end := start + pageSize - 1
		if end > first.TotalCount {
			end = first.TotalCount
		}
		page, err := client.Items(ctx, ecos.ItemsRequest{
			Page:     ecos.NewPage(start, end),
			StatCode: statCode,
		})
		if err != nil {
			return nil, 0, oops.In("collect_series_catalog_inventory").
				With("service", "StatisticItemList", "stat_code", string(statCode), "start", start, "end", end).
				Wrap(err)
		}
		items = append(items, page.Rows...)
	}
	return items, first.TotalCount, nil
}

func toCatalogTable(table ecos.StatisticTable) catalogTable {
	return catalogTable{
		StatCode:       string(table.StatCode),
		ParentStatCode: string(table.ParentStatCode),
		StatName:       table.StatName,
		Cycle:          string(table.Cycle),
		Searchable:     table.Searchable == ecos.SearchableYes,
		Organization:   table.Organization,
	}
}

func toCatalogItem(item ecos.StatisticItem) catalogItem {
	return catalogItem{
		StatCode:       string(item.StatCode),
		StatName:       item.StatName,
		GroupCode:      item.GroupCode,
		GroupName:      item.GroupName,
		ItemCode:       string(item.ItemCode),
		ItemName:       item.ItemName,
		ParentItemCode: string(item.ParentItemCode),
		ParentItemName: item.ParentItemName,
		Cycle:          string(item.Cycle),
		StartTime:      item.StartTime,
		EndTime:        item.EndTime,
		DataCount:      item.DataCount,
		UnitName:       item.UnitName,
		Weight:         item.Weight,
	}
}

func toCatalogKeyStatistic(row ecos.KeyStatistic) catalogKeyStatistic {
	return catalogKeyStatistic{
		ClassName:     row.ClassName,
		Name:          row.Name,
		Value:         row.Value,
		ReferenceTime: row.ReferenceTime,
		UnitName:      row.UnitName,
	}
}

func buildSeries(items []catalogItem) []catalogSeries {
	byStatCycle := map[string][]catalogItem{}
	for _, item := range items {
		if !isSeriesItem(item) {
			continue
		}
		key := strings.Join([]string{item.StatCode, item.Cycle}, ":")
		byStatCycle[key] = append(byStatCycle[key], item)
	}

	keys := make([]string, 0, len(byStatCycle))
	for key := range byStatCycle {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	series := []catalogSeries{}
	for _, key := range keys {
		grouped := groupItemsByPosition(byStatCycle[key])
		positions := sortedGroupPositions(grouped)
		combinations := reviewedCombinations(grouped, positions)
		for _, combination := range combinations {
			series = append(series, toCatalogSeries(combination))
		}
	}
	return series
}

func isSeriesItem(item catalogItem) bool {
	if strings.TrimSpace(item.StatCode) == "" || strings.TrimSpace(item.Cycle) == "" {
		return false
	}
	if strings.TrimSpace(item.ItemCode) == "" {
		return false
	}
	return strings.TrimSpace(item.StartTime) != "" || strings.TrimSpace(item.EndTime) != "" || item.DataCount > 0
}

func groupItemsByPosition(items []catalogItem) map[int][]catalogItem {
	groups := map[int][]catalogItem{}
	fallbackPositions := map[string]int{}
	for _, item := range items {
		position := groupPosition(item.GroupCode, fallbackPositions)
		groups[position] = append(groups[position], item)
	}
	for position := range groups {
		sort.Slice(groups[position], func(left int, right int) bool {
			return groups[position][left].ItemCode < groups[position][right].ItemCode
		})
	}
	return groups
}

func groupPosition(groupCode string, fallbackPositions map[string]int) int {
	normalized := strings.TrimSpace(groupCode)
	matches := groupCodePattern.FindStringSubmatch(normalized)
	if len(matches) == 2 {
		value, err := strconv.Atoi(matches[1])
		if err == nil {
			return value
		}
	}
	if normalized == "" {
		normalized = "Group1"
	}
	if value, ok := fallbackPositions[normalized]; ok {
		return value
	}
	position := len(fallbackPositions) + 1
	if position > 4 {
		position = 4
	}
	fallbackPositions[normalized] = position
	return position
}

func sortedGroupPositions(groups map[int][]catalogItem) []int {
	positions := make([]int, 0, len(groups))
	for position := range groups {
		positions = append(positions, position)
	}
	sort.Ints(positions)
	if len(positions) > 4 {
		return positions[:4]
	}
	return positions
}

func combineGroups(groups map[int][]catalogItem, positions []int) [][]catalogItem {
	if len(positions) == 0 {
		return nil
	}
	combinations := [][]catalogItem{{}}
	for _, position := range positions {
		next := [][]catalogItem{}
		for _, prefix := range combinations {
			for _, item := range groups[position] {
				combination := append([]catalogItem{}, prefix...)
				combination = append(combination, item)
				next = append(next, combination)
			}
		}
		combinations = next
	}
	return combinations
}

func reviewedCombinations(groups map[int][]catalogItem, positions []int) [][]catalogItem {
	if len(positions) <= 1 {
		return combineGroups(groups, positions)
	}

	total := 1
	for _, position := range positions {
		total *= len(groups[position])
		if total > 1 {
			return nil
		}
	}
	return combineGroups(groups, positions)
}

func toCatalogSeries(items []catalogItem) catalogSeries {
	if len(items) == 0 {
		return catalogSeries{}
	}
	itemCodes := make([]string, 0, len(items))
	for _, item := range items {
		itemCodes = append(itemCodes, item.ItemCode)
	}
	status := "queryable"
	if len(items) > 1 {
		status = "needs_review"
	}
	return catalogSeries{
		SeriesKey: strings.Join(append([]string{items[0].StatCode, items[0].Cycle}, itemCodes...), ":"),
		StatCode:  items[0].StatCode,
		Cycle:     items[0].Cycle,
		ItemCodes: itemCodes,
		UnitName:  firstNonEmpty(items, func(item catalogItem) string { return item.UnitName }),
		StartTime: firstNonEmpty(items, func(item catalogItem) string { return item.StartTime }),
		EndTime:   firstNonEmpty(items, func(item catalogItem) string { return item.EndTime }),
		Status:    status,
	}
}

func firstNonEmpty(items []catalogItem, value func(catalogItem) string) string {
	for _, item := range items {
		if trimmed := strings.TrimSpace(value(item)); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func resolveAPIKey(flagValue string, envFile string) (string, error) {
	if strings.TrimSpace(flagValue) != "" {
		return strings.TrimSpace(flagValue), nil
	}
	if value := strings.TrimSpace(os.Getenv("ECOS_API_KEY")); value != "" {
		return value, nil
	}
	return readAPIKeyFromEnvFile(envFile)
}

func readAPIKeyFromEnvFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", oops.In("collect_series_catalog_inventory").
			With("env_file", path).
			Wrap(err)
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		line = strings.TrimPrefix(line, "export ")
		if strings.HasPrefix(line, "#") || !strings.HasPrefix(line, "ECOS_API_KEY=") {
			continue
		}
		value := strings.TrimPrefix(line, "ECOS_API_KEY=")
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		return value, nil
	}
	if err := scanner.Err(); err != nil {
		return "", oops.In("collect_series_catalog_inventory").
			With("env_file", path).
			Wrap(err)
	}
	return "", nil
}

func writeInventory(path string, stdout io.Writer, result inventory) error {
	if isFileOutput(path) {
		return writeInventoryFile(path, result)
	}

	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		return oops.In("collect_series_catalog_inventory").Wrap(err)
	}
	return nil
}

func persistPartialSnapshot(path string, result inventory) error {
	if !isFileOutput(path) {
		return nil
	}
	return writeInventoryFile(path, result)
}

func applyResumeSnapshot(path string, startStatCode string, tables []ecos.StatisticTable, result *inventory) error {
	if !isFileOutput(path) || strings.TrimSpace(startStatCode) == "" {
		return nil
	}

	existing, err := readInventoryFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	previousStatCodes := statCodesBefore(tables, strings.TrimSpace(startStatCode))
	result.Items = filterItemsByStatCode(existing.Items, previousStatCodes)
	result.Series = filterSeriesByStatCode(existing.Series, previousStatCodes)
	result.Summary.ItemTableTotal = countItemTables(result.Items)
	result.Summary.ItemTotal = len(result.Items)
	result.Summary.SeriesTotal = len(result.Series)
	return nil
}

func readInventoryFile(path string) (inventory, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return inventory{}, err
		}
		return inventory{}, oops.In("collect_series_catalog_inventory").
			With("out", path).
			Wrap(err)
	}

	var result inventory
	if err := json.Unmarshal(body, &result); err != nil {
		return inventory{}, oops.In("collect_series_catalog_inventory").
			With("out", path).
			Wrap(err)
	}
	return result, nil
}

func statCodesBefore(tables []ecos.StatisticTable, startStatCode string) map[string]bool {
	statCodes := map[string]bool{}
	for _, table := range tables {
		if table.Searchable != ecos.SearchableYes {
			continue
		}
		if string(table.StatCode) == startStatCode {
			break
		}
		statCodes[string(table.StatCode)] = true
	}
	return statCodes
}

func filterItemsByStatCode(items []catalogItem, statCodes map[string]bool) []catalogItem {
	filtered := make([]catalogItem, 0, len(items))
	for _, item := range items {
		if statCodes[item.StatCode] {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func filterSeriesByStatCode(series []catalogSeries, statCodes map[string]bool) []catalogSeries {
	filtered := make([]catalogSeries, 0, len(series))
	for _, value := range series {
		if statCodes[value.StatCode] {
			filtered = append(filtered, value)
		}
	}
	return filtered
}

func countItemTables(items []catalogItem) int {
	statCodes := map[string]bool{}
	for _, item := range items {
		if strings.TrimSpace(item.StatCode) != "" {
			statCodes[item.StatCode] = true
		}
	}
	return len(statCodes)
}

func isFileOutput(path string) bool {
	return strings.TrimSpace(path) != "" && path != "-"
}

func writeInventoryFile(path string, result inventory) error {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	file, err := os.CreateTemp(dir, "."+base+".*.tmp")
	if err != nil {
		return oops.In("collect_series_catalog_inventory").
			With("out", path).
			Wrap(err)
	}
	tempPath := file.Name()
	shouldRemove := true
	defer func() {
		if shouldRemove {
			_ = os.Remove(tempPath)
		}
	}()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		_ = file.Close()
		return oops.In("collect_series_catalog_inventory").
			With("out", path).
			Wrap(err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return oops.In("collect_series_catalog_inventory").
			With("out", path).
			Wrap(err)
	}
	if err := file.Close(); err != nil {
		return oops.In("collect_series_catalog_inventory").
			With("out", path).
			Wrap(err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return oops.In("collect_series_catalog_inventory").
			With("out", path).
			Wrap(err)
	}
	shouldRemove = false
	return nil
}
