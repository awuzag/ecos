package ecos

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/samber/oops"
)

const (
	serviceStatisticTableList = "StatisticTableList"
	serviceStatisticItemList  = "StatisticItemList"
	serviceStatisticSearch    = "StatisticSearch"
	serviceKeyStatisticList   = "KeyStatisticList"
	serviceStatisticMeta      = "StatisticMeta"
	serviceStatisticWord      = "StatisticWord"

	unusedItemCodeSegment = "?"
)

// PageResult is a typed ECOS paged response.
type PageResult[T any] struct {
	TotalCount int
	RowCount   int
	Rows       []T
}

// TablesRequest holds parameters for StatisticTableList.
type TablesRequest struct {
	Page     Page
	StatCode StatCode
}

// ItemsRequest holds parameters for StatisticItemList.
type ItemsRequest struct {
	Page     Page
	StatCode StatCode
}

// SearchRequest holds parameters for StatisticSearch.
type SearchRequest struct {
	Page      Page
	StatCode  StatCode
	Cycle     Cycle
	StartTime string
	EndTime   string
	ItemCodes []ItemCode
}

// KeyStatisticsRequest holds parameters for KeyStatisticList.
type KeyStatisticsRequest struct {
	Page Page
}

// MetaRequest holds parameters for StatisticMeta.
type MetaRequest struct {
	Page     Page
	DataName string
}

// WordsRequest holds parameters for StatisticWord.
type WordsRequest struct {
	Page Page
	Word string
}

// StatisticTable is a row returned by StatisticTableList.
type StatisticTable struct {
	ParentStatCode StatCode   `json:"P_STAT_CODE"`
	StatCode       StatCode   `json:"STAT_CODE"`
	StatName       string     `json:"STAT_NAME"`
	Cycle          Cycle      `json:"CYCLE"`
	Searchable     SearchFlag `json:"SRCH_YN"`
	Organization   string     `json:"ORG_NAME"`
}

// StatisticItem is a row returned by StatisticItemList.
type StatisticItem struct {
	StatCode       StatCode `json:"STAT_CODE"`
	StatName       string   `json:"STAT_NAME"`
	GroupCode      string   `json:"GRP_CODE"`
	GroupName      string   `json:"GRP_NAME"`
	ItemCode       ItemCode `json:"ITEM_CODE"`
	ItemName       string   `json:"ITEM_NAME"`
	ParentItemCode ItemCode `json:"P_ITEM_CODE"`
	ParentItemName string   `json:"P_ITEM_NAME"`
	Cycle          Cycle    `json:"CYCLE"`
	StartTime      string   `json:"START_TIME"`
	EndTime        string   `json:"END_TIME"`
	DataCount      int      `json:"DATA_CNT"`
	UnitName       string   `json:"UNIT_NAME"`
	Weight         string   `json:"WEIGHT"`
}

// Observation is a row returned by StatisticSearch.
type Observation struct {
	StatCode  StatCode `json:"STAT_CODE"`
	StatName  string   `json:"STAT_NAME"`
	ItemCode1 ItemCode `json:"ITEM_CODE1"`
	ItemName1 string   `json:"ITEM_NAME1"`
	ItemCode2 ItemCode `json:"ITEM_CODE2"`
	ItemName2 string   `json:"ITEM_NAME2"`
	ItemCode3 ItemCode `json:"ITEM_CODE3"`
	ItemName3 string   `json:"ITEM_NAME3"`
	ItemCode4 ItemCode `json:"ITEM_CODE4"`
	ItemName4 string   `json:"ITEM_NAME4"`
	UnitName  string   `json:"UNIT_NAME"`
	Weight    string   `json:"WGT"`
	Time      string   `json:"TIME"`
	Value     string   `json:"DATA_VALUE"`
}

// KeyStatistic is a row returned by KeyStatisticList.
type KeyStatistic struct {
	ClassName     string `json:"CLASS_NAME"`
	Name          string `json:"KEYSTAT_NAME"`
	Value         string `json:"DATA_VALUE"`
	ReferenceTime string `json:"CYCLE"`
	UnitName      string `json:"UNIT_NAME"`
}

// Metadata is a row returned by StatisticMeta.
type Metadata struct {
	Level             string `json:"LVL"`
	ParentContentCode string `json:"P_CONT_CODE"`
	ContentCode       string `json:"CONT_CODE"`
	ContentName       string `json:"CONT_NAME"`
	Data              string `json:"META_DATA"`
}

// Word is a row returned by StatisticWord.
type Word struct {
	Word    string `json:"WORD"`
	Content string `json:"CONTENT"`
}

// Tables calls the ECOS StatisticTableList API.
func (client *Client) Tables(ctx context.Context, req TablesRequest) (PageResult[StatisticTable], error) {
	segments, err := req.Page.pathSegments(serviceStatisticTableList)
	if err != nil {
		return PageResult[StatisticTable]{}, err
	}
	if strings.TrimSpace(string(req.StatCode)) != "" {
		segments = append(segments, string(req.StatCode))
	}
	return getPage[StatisticTable](ctx, client, serviceStatisticTableList, segments)
}

// Items calls the ECOS StatisticItemList API.
func (client *Client) Items(ctx context.Context, req ItemsRequest) (PageResult[StatisticItem], error) {
	if err := requiredString(serviceStatisticItemList, "stat_code", string(req.StatCode)); err != nil {
		return PageResult[StatisticItem]{}, err
	}
	segments, err := req.Page.pathSegments(serviceStatisticItemList)
	if err != nil {
		return PageResult[StatisticItem]{}, err
	}
	segments = append(segments, string(req.StatCode))
	return getPage[StatisticItem](ctx, client, serviceStatisticItemList, segments)
}

// Search calls the ECOS StatisticSearch API.
func (client *Client) Search(ctx context.Context, req SearchRequest) (PageResult[Observation], error) {
	if err := req.validate(); err != nil {
		return PageResult[Observation]{}, err
	}
	segments, err := req.Page.pathSegments(serviceStatisticSearch)
	if err != nil {
		return PageResult[Observation]{}, err
	}
	segments = append(
		segments,
		string(req.StatCode),
		string(req.Cycle),
		req.StartTime,
		req.EndTime,
	)
	segments = append(segments, searchItemSegments(req.ItemCodes)...)
	return getPage[Observation](ctx, client, serviceStatisticSearch, segments)
}

// KeyStatistics calls the ECOS KeyStatisticList API.
func (client *Client) KeyStatistics(ctx context.Context, req KeyStatisticsRequest) (PageResult[KeyStatistic], error) {
	segments, err := req.Page.pathSegments(serviceKeyStatisticList)
	if err != nil {
		return PageResult[KeyStatistic]{}, err
	}
	return getPage[KeyStatistic](ctx, client, serviceKeyStatisticList, segments)
}

// Meta calls the ECOS StatisticMeta API.
func (client *Client) Meta(ctx context.Context, req MetaRequest) (PageResult[Metadata], error) {
	if err := requiredString(serviceStatisticMeta, "data_name", req.DataName); err != nil {
		return PageResult[Metadata]{}, err
	}
	segments, err := req.Page.pathSegments(serviceStatisticMeta)
	if err != nil {
		return PageResult[Metadata]{}, err
	}
	segments = append(segments, req.DataName)
	return getPage[Metadata](ctx, client, serviceStatisticMeta, segments)
}

// Words calls the ECOS StatisticWord API.
func (client *Client) Words(ctx context.Context, req WordsRequest) (PageResult[Word], error) {
	if err := requiredString(serviceStatisticWord, "word", req.Word); err != nil {
		return PageResult[Word]{}, err
	}
	segments, err := req.Page.pathSegments(serviceStatisticWord)
	if err != nil {
		return PageResult[Word]{}, err
	}
	segments = append(segments, req.Word)
	return getPage[Word](ctx, client, serviceStatisticWord, segments)
}

func (req SearchRequest) validate() error {
	if err := requiredString(serviceStatisticSearch, "stat_code", string(req.StatCode)); err != nil {
		return err
	}
	if err := validateCycle(serviceStatisticSearch, req.Cycle); err != nil {
		return err
	}
	if err := validateCycleTime(serviceStatisticSearch, "start_time", req.Cycle, req.StartTime); err != nil {
		return err
	}
	if err := validateCycleTime(serviceStatisticSearch, "end_time", req.Cycle, req.EndTime); err != nil {
		return err
	}
	if len(req.ItemCodes) > 4 {
		return validationError(serviceStatisticSearch, "item_codes", "must contain at most 4 values")
	}
	return nil
}

func searchItemSegments(itemCodes []ItemCode) []string {
	segments := make([]string, 4)
	for index := range segments {
		segments[index] = unusedItemCodeSegment
	}
	for index, itemCode := range itemCodes {
		code := strings.TrimSpace(string(itemCode))
		if code != "" {
			segments[index] = code
		}
	}
	return segments
}

type pageEnvelope[T any] struct {
	TotalCount int `json:"list_total_count"`
	RowCount   int `json:"row_count"`
	Rows       []T `json:"row"`
}

func getPage[T any](ctx context.Context, client *Client, service string, pathSegments []string) (PageResult[T], error) {
	body, endpoint, err := getRawJSON(ctx, client, service, pathSegments)
	if err != nil {
		return PageResult[T]{}, err
	}

	var root map[string]json.RawMessage
	if err := json.Unmarshal(body, &root); err != nil {
		return PageResult[T]{}, decodeError("GET", endpoint, service, defaultFormat, err)
	}

	rawPage, ok := root[service]
	if !ok {
		return PageResult[T]{}, oops.In("decode").
			With("method", "GET", "endpoint", endpoint, "op", service, "root", service).
			New("ecos: response root is missing")
	}

	var page pageEnvelope[T]
	if err := json.Unmarshal(rawPage, &page); err != nil {
		return PageResult[T]{}, decodeError("GET", endpoint, service, defaultFormat, err)
	}

	return PageResult[T]{
		TotalCount: page.TotalCount,
		RowCount:   page.RowCount,
		Rows:       page.Rows,
	}, nil
}
