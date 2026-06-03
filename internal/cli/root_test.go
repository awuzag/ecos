package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShowKeyStatisticsJSON(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"KeyStatisticList":{"list_total_count":1,"row_count":1,"row":[{"CLASS_NAME":"시장금리","KEYSTAT_NAME":"한국은행 기준금리","DATA_VALUE":"2.5","CYCLE":"20260528","UNIT_NAME":"%"}]}}`))
	}))
	t.Cleanup(server.Close)

	out, errOut, code := execute(t,
		"show", "key-statistics",
		"--api-key", "sample",
		"--base-url", server.URL,
		"--start", "1",
		"--end", "5",
		"--json",
	)

	require.Equal(t, 0, code, errOut)
	require.Equal(t, "/KeyStatisticList/sample/json/kr/1/5", gotPath)

	var payload struct {
		Rows []struct {
			Name          string `json:"KEYSTAT_NAME"`
			ReferenceTime string `json:"CYCLE"`
		}
	}
	require.NoError(t, json.Unmarshal([]byte(out), &payload))
	require.Len(t, payload.Rows, 1)
	require.Equal(t, "한국은행 기준금리", payload.Rows[0].Name)
	require.Equal(t, "20260528", payload.Rows[0].ReferenceTime)
}

func TestSearchObservationsTextBuildsPath(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"StatisticSearch":{"list_total_count":1,"row":[{"STAT_CODE":"722Y001","STAT_NAME":"한국은행 기준금리","ITEM_CODE1":"0101000","ITEM_NAME1":"한국은행 기준금리","ITEM_CODE2":"","ITEM_NAME2":"","ITEM_CODE3":"","ITEM_NAME3":"","ITEM_CODE4":"","ITEM_NAME4":"","UNIT_NAME":"연%","WGT":"","TIME":"202604","DATA_VALUE":"3.5"}]}}`))
	}))
	t.Cleanup(server.Close)

	out, errOut, code := execute(t,
		"--api-key", "sample",
		"--base-url", server.URL,
		"search", "observations",
		"--stat-code", "722Y001",
		"--cycle", "M",
		"--time-start", "202001",
		"--time-end", "202604",
		"--item-code", "0101000",
	)

	require.Equal(t, 0, code, errOut)
	require.Equal(t, "/StatisticSearch/sample/json/kr/1/10/722Y001/M/202001/202604/0101000/%3F/%3F/%3F", gotPath)
	require.Contains(t, out, "TIME")
	require.Contains(t, out, "202604")
	require.Contains(t, out, "3.5")
}

func TestRemainingVerbFirstCommandsBuildSDKPaths(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		response     string
		expectedPath string
		wantOutput   string
	}{
		{
			name: "list items",
			args: []string{
				"list", "items",
				"--stat-code", "722Y001",
				"--start", "2",
				"--end", "3",
			},
			response:     `{"StatisticItemList":{"list_total_count":1,"row":[{"STAT_CODE":"722Y001","STAT_NAME":"한국은행 기준금리","GRP_CODE":"Group1","GRP_NAME":"항목그룹","ITEM_CODE":"0101000","ITEM_NAME":"한국은행 기준금리","P_ITEM_CODE":"","P_ITEM_NAME":"","CYCLE":"M","START_TIME":"199905","END_TIME":"202604","DATA_CNT":324,"UNIT_NAME":"연%","WEIGHT":""}]}}`,
			expectedPath: "/StatisticItemList/sample/json/kr/2/3/722Y001",
			wantOutput:   "0101000",
		},
		{
			name: "show word",
			args: []string{
				"show", "word", "소비자동향지수",
			},
			response:     `{"StatisticWord":{"list_total_count":1,"row":[{"WORD":"소비자동향지수","CONTENT":"설명"}]}}`,
			expectedPath: "/StatisticWord/sample/json/kr/1/10/" + url.PathEscape("소비자동향지수"),
			wantOutput:   "소비자동향지수",
		},
		{
			name: "show meta",
			args: []string{
				"show", "meta",
				"--data-name", "경제심리지수",
			},
			response:     `{"StatisticMeta":{"list_total_count":1,"row":[{"LVL":"1","P_CONT_CODE":"","CONT_CODE":"1","CONT_NAME":"경제심리지수","META_DATA":"설명"}]}}`,
			expectedPath: "/StatisticMeta/sample/json/kr/1/10/" + url.PathEscape("경제심리지수"),
			wantOutput:   "경제심리지수",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.EscapedPath()
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tt.response))
			}))
			t.Cleanup(server.Close)

			args := append([]string{"--api-key", "sample", "--base-url", server.URL}, tt.args...)
			out, errOut, code := execute(t, args...)

			require.Equal(t, 0, code, errOut)
			require.Equal(t, tt.expectedPath, gotPath)
			require.Contains(t, out, tt.wantOutput)
		})
	}
}

func TestEnvFileAPIKeyIsUsedAndNotPrinted(t *testing.T) {
	t.Setenv(ecosAPIKeyEnv, "")
	envFile := filepath.Join(t.TempDir(), ".env")
	require.NoError(t, os.WriteFile(envFile, []byte("ECOS_API_KEY=from-file\n"), 0o600))

	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"StatisticTableList":{"list_total_count":1,"row":[{"P_STAT_CODE":"000000","STAT_CODE":"722Y001","STAT_NAME":"한국은행 기준금리","CYCLE":"D","SRCH_YN":"Y","ORG_NAME":"한국은행"}]}}`))
	}))
	t.Cleanup(server.Close)

	out, errOut, code := execute(t,
		"list", "tables",
		"--env-file", envFile,
		"--base-url", server.URL,
	)

	require.Equal(t, 0, code, errOut)
	require.Equal(t, "/StatisticTableList/from-file/json/kr/1/10", gotPath)
	require.Contains(t, out, "722Y001")
	require.NotContains(t, out, "from-file")
	require.NotContains(t, errOut, "from-file")
}

func TestAPIKeyFlagPrecedesEnvironmentSources(t *testing.T) {
	t.Setenv(ecosAPIKeyEnv, "from-env")
	envFile := filepath.Join(t.TempDir(), ".env")
	require.NoError(t, os.WriteFile(envFile, []byte("ECOS_API_KEY=from-file\n"), 0o600))

	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"KeyStatisticList":{"list_total_count":0,"row":[]}}`))
	}))
	t.Cleanup(server.Close)

	_, errOut, code := execute(t,
		"--api-key", "from-flag",
		"--env-file", envFile,
		"--base-url", server.URL,
		"show", "key-statistics",
	)

	require.Equal(t, 0, code, errOut)
	require.Equal(t, "/KeyStatisticList/from-flag/json/kr/1/10", gotPath)
}

func execute(t *testing.T, args ...string) (string, string, int) {
	t.Helper()

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Execute(context.Background(), args, &out, &errOut)
	return out.String(), errOut.String(), code
}
