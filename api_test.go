package ecos

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTablesBuildsPathAndDecodesRows(t *testing.T) {
	var gotPath string
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"StatisticTableList":{"list_total_count":1,"row":[{"P_STAT_CODE":"000000","STAT_CODE":"722Y001","STAT_NAME":"한국은행 기준금리","CYCLE":"D","SRCH_YN":"Y","ORG_NAME":"한국은행"}]}}`))
	})

	result, err := client.Tables(context.Background(), TablesRequest{
		Page:     NewPage(1, 1),
		StatCode: StatCodeBankOfKoreaBaseRate,
	})

	require.NoError(t, err)
	require.Equal(t, "/StatisticTableList/sample/json/kr/1/1/722Y001", gotPath)
	require.Equal(t, 1, result.TotalCount)
	require.Len(t, result.Rows, 1)
	require.Equal(t, StatCodeBankOfKoreaBaseRate, result.Rows[0].StatCode)
	require.Equal(t, CycleDaily, result.Rows[0].Cycle)
	require.Equal(t, SearchableYes, result.Rows[0].Searchable)
}

func TestItemsBuildsPathAndDecodesRows(t *testing.T) {
	var gotPath string
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"StatisticItemList":{"list_total_count":1,"row":[{"STAT_CODE":"722Y001","STAT_NAME":"한국은행 기준금리","GRP_CODE":"Group1","GRP_NAME":"항목그룹","ITEM_CODE":"0101000","ITEM_NAME":"한국은행 기준금리","P_ITEM_CODE":"","P_ITEM_NAME":"","CYCLE":"M","START_TIME":"199905","END_TIME":"202604","DATA_CNT":324,"UNIT_NAME":"연%","WEIGHT":""}]}}`))
	})

	result, err := client.Items(context.Background(), ItemsRequest{
		Page:     NewPage(1, 100),
		StatCode: StatCodeBankOfKoreaBaseRate,
	})

	require.NoError(t, err)
	require.Equal(t, "/StatisticItemList/sample/json/kr/1/100/722Y001", gotPath)
	require.Len(t, result.Rows, 1)
	require.Equal(t, ItemCodeBankOfKoreaBaseRate, result.Rows[0].ItemCode)
	require.Equal(t, CycleMonthly, result.Rows[0].Cycle)
	require.Equal(t, "199905", result.Rows[0].StartTime)
	require.Equal(t, 324, result.Rows[0].DataCount)
}

func TestSearchBuildsUnusedItemCodeSegmentsAndDecodesRows(t *testing.T) {
	var gotPath string
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"StatisticSearch":{"list_total_count":1,"row":[{"STAT_CODE":"722Y001","STAT_NAME":"한국은행 기준금리","ITEM_CODE1":"0101000","ITEM_NAME1":"한국은행 기준금리","ITEM_CODE2":"","ITEM_NAME2":"","ITEM_CODE3":"","ITEM_NAME3":"","ITEM_CODE4":"","ITEM_NAME4":"","UNIT_NAME":"연%","WGT":"","TIME":"202604","DATA_VALUE":"3.5"}]}}`))
	})

	result, err := client.Search(context.Background(), SearchRequest{
		Page:      NewPage(1, 10),
		StatCode:  StatCodeBankOfKoreaBaseRate,
		Cycle:     CycleMonthly,
		StartTime: "202001",
		EndTime:   "202604",
		ItemCodes: []ItemCode{ItemCodeBankOfKoreaBaseRate},
	})

	require.NoError(t, err)
	require.Equal(t, "/StatisticSearch/sample/json/kr/1/10/722Y001/M/202001/202604/0101000/%3F/%3F/%3F", gotPath)
	require.Len(t, result.Rows, 1)
	require.Equal(t, "202604", result.Rows[0].Time)
	require.Equal(t, "3.5", result.Rows[0].Value)
}

func TestSearchUsesQuestionMarkForAllMissingItemCodes(t *testing.T) {
	var gotPath string
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"StatisticSearch":{"list_total_count":0,"row":[]}}`))
	})

	_, err := client.Search(context.Background(), SearchRequest{
		Page:      NewPage(1, 10),
		StatCode:  StatCodeBankOfKoreaBaseRate,
		Cycle:     CycleAnnual,
		StartTime: "2020",
		EndTime:   "2024",
	})

	require.NoError(t, err)
	require.Equal(t, "/StatisticSearch/sample/json/kr/1/10/722Y001/A/2020/2024/%3F/%3F/%3F/%3F", gotPath)
}

func TestKeyStatisticsDecodesReferenceTimeFromCycleField(t *testing.T) {
	var gotPath string
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"KeyStatisticList":{"list_total_count":1,"row_count":1,"row":[{"CLASS_NAME":"시장금리","KEYSTAT_NAME":"한국은행 기준금리","DATA_VALUE":"2.5","CYCLE":"20260528","UNIT_NAME":"%"}]}}`))
	})

	result, err := client.KeyStatistics(context.Background(), KeyStatisticsRequest{Page: NewPage(1, 1)})

	require.NoError(t, err)
	require.Equal(t, "/KeyStatisticList/sample/json/kr/1/1", gotPath)
	require.Equal(t, 1, result.RowCount)
	require.Len(t, result.Rows, 1)
	require.Equal(t, "20260528", result.Rows[0].ReferenceTime)
}

func TestMetaEscapesKoreanDataName(t *testing.T) {
	var gotPath string
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"StatisticMeta":{"list_total_count":1,"row":[{"LVL":"1","P_CONT_CODE":"","CONT_CODE":"1","CONT_NAME":"경제심리지수","META_DATA":"설명"}]}}`))
	})

	result, err := client.Meta(context.Background(), MetaRequest{
		Page:     NewPage(1, 10),
		DataName: "경제심리지수",
	})

	require.NoError(t, err)
	require.Equal(t, "/StatisticMeta/sample/json/kr/1/10/"+url.PathEscape("경제심리지수"), gotPath)
	require.Len(t, result.Rows, 1)
	require.Equal(t, "경제심리지수", result.Rows[0].ContentName)
}

func TestWordsEscapesKoreanWord(t *testing.T) {
	var gotPath string
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"StatisticWord":{"list_total_count":1,"row":[{"WORD":"소비자동향지수","CONTENT":"설명"}]}}`))
	})

	result, err := client.Words(context.Background(), WordsRequest{
		Page: NewPage(1, 10),
		Word: "소비자동향지수",
	})

	require.NoError(t, err)
	require.Equal(t, "/StatisticWord/sample/json/kr/1/10/"+url.PathEscape("소비자동향지수"), gotPath)
	require.Len(t, result.Rows, 1)
	require.Equal(t, "소비자동향지수", result.Rows[0].Word)
}

func TestRequestValidation(t *testing.T) {
	client, err := New(Config{APIKey: "sample"})
	require.NoError(t, err)

	_, err = client.KeyStatistics(context.Background(), KeyStatisticsRequest{Page: NewPage(0, 1)})
	require.Error(t, err)
	require.Contains(t, err.Error(), "page.start")

	_, err = client.KeyStatistics(context.Background(), KeyStatisticsRequest{Page: NewPage(2, 1)})
	require.Error(t, err)
	require.Contains(t, err.Error(), "page.end")

	_, err = client.Items(context.Background(), ItemsRequest{Page: NewPage(1, 1)})
	require.Error(t, err)
	require.Contains(t, err.Error(), "stat_code")

	_, err = client.Meta(context.Background(), MetaRequest{Page: NewPage(1, 1)})
	require.Error(t, err)
	require.Contains(t, err.Error(), "data_name")

	_, err = client.Words(context.Background(), WordsRequest{Page: NewPage(1, 1)})
	require.Error(t, err)
	require.Contains(t, err.Error(), "word")

	_, err = client.Search(context.Background(), SearchRequest{
		Page:      NewPage(1, 1),
		StatCode:  StatCodeBankOfKoreaBaseRate,
		Cycle:     CycleMonthly,
		StartTime: "2020",
		EndTime:   "202001",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "start_time")

	_, err = client.Search(context.Background(), SearchRequest{
		Page:      NewPage(1, 1),
		StatCode:  StatCodeBankOfKoreaBaseRate,
		Cycle:     Cycle("W"),
		StartTime: "202001",
		EndTime:   "202001",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "cycle")

	_, err = client.Search(context.Background(), SearchRequest{
		Page:      NewPage(1, 1),
		StatCode:  StatCodeBankOfKoreaBaseRate,
		Cycle:     CycleMonthly,
		StartTime: "202001",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "end_time")

	_, err = client.Search(context.Background(), SearchRequest{
		Page:      NewPage(1, 1),
		StatCode:  StatCodeBankOfKoreaBaseRate,
		Cycle:     CycleMonthly,
		StartTime: "202001",
		EndTime:   "202002",
		ItemCodes: []ItemCode{"1", "2", "3", "4", "5"},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "item_codes")
}

func TestCycleValidTime(t *testing.T) {
	require.True(t, CycleAnnual.ValidTime("2024"))
	require.True(t, CycleSemiAnnual.ValidTime("2024S2"))
	require.True(t, CycleQuarterly.ValidTime("2024Q4"))
	require.True(t, CycleMonthly.ValidTime("202412"))
	require.True(t, CycleSemiMonthly.ValidTime("202412S1"))
	require.True(t, CycleDaily.ValidTime("20241231"))

	require.False(t, CycleAnnual.ValidTime("202401"))
	require.False(t, CycleSemiAnnual.ValidTime("2024S3"))
	require.False(t, CycleQuarterly.ValidTime("2024Q5"))
	require.False(t, CycleMonthly.ValidTime("202413"))
	require.False(t, CycleSemiMonthly.ValidTime("202400S1"))
	require.False(t, CycleDaily.ValidTime("20240232"))
	require.False(t, Cycle("W").ValidTime("20240101"))
}

func TestGetPageReturnsDecodeErrorForMissingRoot(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"OtherService":{"list_total_count":0,"row":[]}}`))
	})

	_, err := client.KeyStatistics(context.Background(), KeyStatisticsRequest{Page: NewPage(1, 1)})

	require.Error(t, err)
	require.Contains(t, err.Error(), "response root")
}

func TestGetPageReturnsDecodeErrorForInvalidServicePayload(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"KeyStatisticList":{"list_total_count":1,"row":{}}}`))
	})

	_, err := client.KeyStatistics(context.Background(), KeyStatisticsRequest{Page: NewPage(1, 1)})

	require.Error(t, err)
	var decodeErr *DecodeError
	require.ErrorAs(t, err, &decodeErr)
	require.Equal(t, serviceKeyStatisticList, decodeErr.Op)
}

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client, err := New(Config{APIKey: "sample"}, WithBaseURL(server.URL))
	require.NoError(t, err)
	return client
}
