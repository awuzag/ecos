package cli

import (
	"context"
	"encoding/json"
	"io"
	"strings"

	"github.com/awuzag/ecos"
	"github.com/samber/oops"
	"github.com/spf13/cobra"
)

type rootOptions struct {
	apiKey  string
	envFile string
	baseURL string
	start   int
	end     int
	json    bool
}

type runner struct {
	out     io.Writer
	errOut  io.Writer
	options rootOptions
}

// Execute runs the ecos CLI and returns a process exit code.
func Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) int {
	cmd := NewRootCommand(out, errOut)
	cmd.SetArgs(args)
	if err := cmd.ExecuteContext(ctx); err != nil {
		_, _ = errOut.Write([]byte(err.Error() + "\n"))
		return 1
	}
	return 0
}

// NewRootCommand builds the ecos Cobra command tree.
func NewRootCommand(out io.Writer, errOut io.Writer) *cobra.Command {
	r := &runner{
		out:    out,
		errOut: errOut,
		options: rootOptions{
			envFile: ".env",
			start:   1,
			end:     10,
		},
	}

	root := &cobra.Command{
		Use:           "ecos",
		Short:         "한국은행 ECOS OpenAPI CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetOut(out)
	root.SetErr(errOut)
	root.PersistentFlags().StringVar(&r.options.apiKey, "api-key", "", "ECOS API key")
	root.PersistentFlags().StringVar(&r.options.envFile, "env-file", ".env", "env file path for ECOS_API_KEY")
	root.PersistentFlags().StringVar(&r.options.baseURL, "base-url", "", "ECOS API base URL")
	root.PersistentFlags().IntVar(&r.options.start, "start", 1, "ECOS start row")
	root.PersistentFlags().IntVar(&r.options.end, "end", 10, "ECOS end row")
	root.PersistentFlags().BoolVar(&r.options.json, "json", false, "print JSON output")

	root.AddCommand(r.newListCommand())
	root.AddCommand(r.newSearchCommand())
	root.AddCommand(r.newShowCommand())
	return root
}

func (r *runner) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "목록 조회",
		Args:  noArgs("list"),
	}
	cmd.AddCommand(r.newListTablesCommand())
	cmd.AddCommand(r.newListItemsCommand())
	return cmd
}

func (r *runner) newListTablesCommand() *cobra.Command {
	var statCode string
	cmd := &cobra.Command{
		Use:   "tables",
		Short: "통계표 목록 조회",
		Args:  noArgs("tables"),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := r.client(cmd)
			if err != nil {
				return err
			}
			result, err := client.Tables(cmd.Context(), ecos.TablesRequest{
				Page:     r.page(),
				StatCode: ecos.StatCode(strings.TrimSpace(statCode)),
			})
			if err != nil {
				return err
			}
			return r.writeResult(result, tableRows(
				[]string{"STAT_CODE", "STAT_NAME", "CYCLE", "SRCH_YN", "ORG_NAME"},
				result.Rows,
				func(row ecos.StatisticTable) []string {
					return []string{string(row.StatCode), row.StatName, string(row.Cycle), string(row.Searchable), row.Organization}
				},
			))
		},
	}
	cmd.Flags().StringVar(&statCode, "stat-code", "", "filter by statistic table code")
	return cmd
}

func (r *runner) newListItemsCommand() *cobra.Command {
	var statCode string
	cmd := &cobra.Command{
		Use:   "items",
		Short: "통계항목 목록 조회",
		Args:  noArgs("items"),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := r.client(cmd)
			if err != nil {
				return err
			}
			result, err := client.Items(cmd.Context(), ecos.ItemsRequest{
				Page:     r.page(),
				StatCode: ecos.StatCode(strings.TrimSpace(statCode)),
			})
			if err != nil {
				return err
			}
			return r.writeResult(result, tableRows(
				[]string{"STAT_CODE", "ITEM_CODE", "ITEM_NAME", "CYCLE", "START_TIME", "END_TIME", "UNIT"},
				result.Rows,
				func(row ecos.StatisticItem) []string {
					return []string{string(row.StatCode), string(row.ItemCode), row.ItemName, string(row.Cycle), row.StartTime, row.EndTime, row.UnitName}
				},
			))
		},
	}
	cmd.Flags().StringVar(&statCode, "stat-code", "", "statistic table code")
	return cmd
}

func (r *runner) newSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "통계 데이터 검색",
		Args:  noArgs("search"),
	}
	cmd.AddCommand(r.newSearchObservationsCommand())
	return cmd
}

func (r *runner) newSearchObservationsCommand() *cobra.Command {
	var statCode string
	var cycle string
	var startTime string
	var endTime string
	var itemCodes []string
	cmd := &cobra.Command{
		Use:   "observations",
		Short: "통계 시계열 조회",
		Args:  noArgs("observations"),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := r.client(cmd)
			if err != nil {
				return err
			}
			result, err := client.Search(cmd.Context(), ecos.SearchRequest{
				Page:      r.page(),
				StatCode:  ecos.StatCode(strings.TrimSpace(statCode)),
				Cycle:     ecos.Cycle(strings.TrimSpace(cycle)),
				StartTime: strings.TrimSpace(startTime),
				EndTime:   strings.TrimSpace(endTime),
				ItemCodes: toItemCodes(itemCodes),
			})
			if err != nil {
				return err
			}
			return r.writeResult(result, tableRows(
				[]string{"TIME", "VALUE", "STAT_CODE", "ITEM_CODE", "ITEM_NAME", "UNIT"},
				result.Rows,
				func(row ecos.Observation) []string {
					return []string{row.Time, row.Value, string(row.StatCode), string(row.ItemCode1), row.ItemName1, row.UnitName}
				},
			))
		},
	}
	cmd.Flags().StringVar(&statCode, "stat-code", "", "statistic table code")
	cmd.Flags().StringVar(&cycle, "cycle", "", "ECOS cycle code")
	cmd.Flags().StringVar(&startTime, "time-start", "", "ECOS period start")
	cmd.Flags().StringVar(&endTime, "time-end", "", "ECOS period end")
	cmd.Flags().StringArrayVar(&itemCodes, "item-code", nil, "ECOS item code, repeatable up to 4")
	return cmd
}

func (r *runner) newShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "단건 성격 데이터 조회",
		Args:  noArgs("show"),
	}
	cmd.AddCommand(r.newShowKeyStatisticsCommand())
	cmd.AddCommand(r.newShowWordCommand())
	cmd.AddCommand(r.newShowMetaCommand())
	return cmd
}

func (r *runner) newShowKeyStatisticsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "key-statistics",
		Short: "주요 지표 조회",
		Args:  noArgs("key-statistics"),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := r.client(cmd)
			if err != nil {
				return err
			}
			result, err := client.KeyStatistics(cmd.Context(), ecos.KeyStatisticsRequest{Page: r.page()})
			if err != nil {
				return err
			}
			return r.writeResult(result, tableRows(
				[]string{"CLASS", "NAME", "VALUE", "REFERENCE_TIME", "UNIT"},
				result.Rows,
				func(row ecos.KeyStatistic) []string {
					return []string{row.ClassName, row.Name, row.Value, row.ReferenceTime, row.UnitName}
				},
			))
		},
	}
}

func (r *runner) newShowWordCommand() *cobra.Command {
	var word string
	cmd := &cobra.Command{
		Use:   "word [word]",
		Short: "통계용어 조회",
		Args:  optionalSingleArg("word"),
		RunE: func(cmd *cobra.Command, args []string) error {
			value, err := singleValue("word", word, args)
			if err != nil {
				return err
			}
			client, err := r.client(cmd)
			if err != nil {
				return err
			}
			result, err := client.Words(cmd.Context(), ecos.WordsRequest{
				Page: r.page(),
				Word: value,
			})
			if err != nil {
				return err
			}
			return r.writeResult(result, tableRows(
				[]string{"WORD", "CONTENT"},
				result.Rows,
				func(row ecos.Word) []string {
					return []string{row.Word, compactText(row.Content)}
				},
			))
		},
	}
	cmd.Flags().StringVar(&word, "word", "", "ECOS dictionary word")
	return cmd
}

func (r *runner) newShowMetaCommand() *cobra.Command {
	var dataName string
	cmd := &cobra.Command{
		Use:   "meta [data-name]",
		Short: "통계 메타데이터 조회",
		Args:  optionalSingleArg("meta"),
		RunE: func(cmd *cobra.Command, args []string) error {
			value, err := singleValue("data_name", dataName, args)
			if err != nil {
				return err
			}
			client, err := r.client(cmd)
			if err != nil {
				return err
			}
			result, err := client.Meta(cmd.Context(), ecos.MetaRequest{
				Page:     r.page(),
				DataName: value,
			})
			if err != nil {
				return err
			}
			return r.writeResult(result, tableRows(
				[]string{"CONT_CODE", "CONT_NAME", "META_DATA"},
				result.Rows,
				func(row ecos.Metadata) []string {
					return []string{row.ContentCode, row.ContentName, compactText(row.Data)}
				},
			))
		},
	}
	cmd.Flags().StringVar(&dataName, "data-name", "", "ECOS metadata name")
	return cmd
}

func (r *runner) client(cmd *cobra.Command) (*ecos.Client, error) {
	apiKey, err := resolveAPIKey(r.options.apiKey, r.options.envFile, cmd.Root().PersistentFlags().Changed("env-file"))
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(apiKey) == "" {
		return nil, oops.In("cli").
			With("field", "api_key").
			New("ecos: API key is required; set --api-key, ECOS_API_KEY, or repo-local .env")
	}

	opts := make([]ecos.Option, 0, 1)
	if strings.TrimSpace(r.options.baseURL) != "" {
		opts = append(opts, ecos.WithBaseURL(r.options.baseURL))
	}
	return ecos.New(ecos.Config{APIKey: apiKey}, opts...)
}

func (r *runner) page() ecos.Page {
	return ecos.NewPage(r.options.start, r.options.end)
}

func (r *runner) writeResult(value any, text string) error {
	if r.options.json {
		encoder := json.NewEncoder(r.out)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(value); err != nil {
			return oops.In("cli").
				With("format", "json").
				Wrap(err)
		}
		return nil
	}
	_, err := io.WriteString(r.out, text)
	if err != nil {
		return oops.In("cli").
			With("format", "text").
			Wrap(err)
	}
	return nil
}

func toItemCodes(values []string) []ecos.ItemCode {
	codes := make([]ecos.ItemCode, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		codes = append(codes, ecos.ItemCode(trimmed))
	}
	return codes
}

func noArgs(name string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return nil
		}
		return oops.In("cli").
			With("command", name).
			Errorf("ecos: %s does not accept positional arguments", name)
	}
}

func optionalSingleArg(name string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) <= 1 {
			return nil
		}
		return oops.In("cli").
			With("command", name).
			Errorf("ecos: %s accepts at most one positional argument", name)
	}
}

func singleValue(field string, flagValue string, args []string) (string, error) {
	value := strings.TrimSpace(flagValue)
	if len(args) == 1 {
		arg := strings.TrimSpace(args[0])
		if value != "" && arg != "" {
			return "", oops.In("cli").
				With("field", field).
				Errorf("ecos: pass %s either as a flag or an argument, not both", field)
		}
		value = arg
	}
	return value, nil
}
