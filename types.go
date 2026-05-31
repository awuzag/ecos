package ecos

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/samber/oops"
)

// Format is an ECOS response format code.
type Format string

const (
	// FormatJSON requests JSON responses.
	FormatJSON Format = "json"
	// FormatXML requests XML responses.
	FormatXML Format = "xml"
)

// Lang is an ECOS response language code.
type Lang string

const (
	// LangKorean requests Korean responses.
	LangKorean Lang = "kr"
	// LangEnglish requests English responses.
	LangEnglish Lang = "en"
)

// MessageCode is a provider-native ECOS API message code.
type MessageCode string

const (
	MessageInvalidAPIKey     MessageCode = "APIMSG001-100"
	MessageNoData            MessageCode = "APIMSG001-200"
	MessageMissingRequired   MessageCode = "APIMSG002-100"
	MessageInvalidCycleTime  MessageCode = "APIMSG002-101"
	MessageInvalidFormat     MessageCode = "APIMSG002-200"
	MessageMissingPage       MessageCode = "APIMSG002-300"
	MessageInvalidPage       MessageCode = "APIMSG002-301"
	MessageSearchTimeout     MessageCode = "APIMSG002-400"
	MessageServerError       MessageCode = "APIMSG002-500"
	MessageDBConnectionError MessageCode = "APIMSG002-600"
	MessageSQLError          MessageCode = "APIMSG002-601"
	MessageTooManyRequests   MessageCode = "APIMSG002-602"
)

// Cycle is an ECOS statistic period code.
type Cycle string

const (
	// CycleAnnual is the annual period code.
	CycleAnnual Cycle = "A"
	// CycleSemiAnnual is the half-year period code.
	CycleSemiAnnual Cycle = "S"
	// CycleQuarterly is the quarterly period code.
	CycleQuarterly Cycle = "Q"
	// CycleMonthly is the monthly period code.
	CycleMonthly Cycle = "M"
	// CycleSemiMonthly is the half-month period code.
	CycleSemiMonthly Cycle = "SM"
	// CycleDaily is the daily period code.
	CycleDaily Cycle = "D"
)

// StatCode is an ECOS statistic table code.
type StatCode string

const (
	// StatCodeBankOfKoreaBaseRate is the 한국은행 기준금리 table used in official examples.
	StatCodeBankOfKoreaBaseRate StatCode = "722Y001"
	// StatCodeDailyMarketInterestRate is the daily market interest rate table used in official examples.
	StatCodeDailyMarketInterestRate StatCode = "817Y002"
	// StatCodeMonetaryBase is the monetary base table used in official examples.
	StatCodeMonetaryBase StatCode = "102Y001"
)

// ItemCode is an ECOS statistic item code.
type ItemCode string

const (
	// ItemCodeBankOfKoreaBaseRate is the 한국은행 기준금리 item used in official examples.
	ItemCodeBankOfKoreaBaseRate ItemCode = "0101000"
	// ItemCodeKORIBOR3M is the KORIBOR 3-month item used in local research notes.
	ItemCodeKORIBOR3M ItemCode = "010150000"
	// ItemCodeCD91D is the CD 91-day item used in local research notes.
	ItemCodeCD91D ItemCode = "010502000"
	// ItemCodeMonetaryStabilizationBond1Y is the one-year MSB item used in local research notes.
	ItemCodeMonetaryStabilizationBond1Y ItemCode = "010400001"
)

// SearchFlag is the provider-native Y/N flag for searchable statistic tables.
type SearchFlag string

const (
	// SearchableYes means the statistic table can be queried by StatisticSearch.
	SearchableYes SearchFlag = "Y"
	// SearchableNo means the statistic table is not directly queryable.
	SearchableNo SearchFlag = "N"
)

// Page selects ECOS path-style start and end row numbers.
type Page struct {
	Start int
	End   int
}

// NewPage creates a Page for ECOS path-style paging.
func NewPage(start int, end int) Page {
	return Page{Start: start, End: end}
}

var cycleTimePatterns = map[Cycle]*regexp.Regexp{
	CycleAnnual:      regexp.MustCompile(`^\d{4}$`),
	CycleSemiAnnual:  regexp.MustCompile(`^\d{4}S[12]$`),
	CycleQuarterly:   regexp.MustCompile(`^\d{4}Q[1-4]$`),
	CycleMonthly:     regexp.MustCompile(`^\d{4}(0[1-9]|1[0-2])$`),
	CycleSemiMonthly: regexp.MustCompile(`^\d{4}(0[1-9]|1[0-2])S[12]$`),
	CycleDaily:       regexp.MustCompile(`^\d{4}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])$`),
}

// ValidTime reports whether value matches the ECOS time format for the cycle.
func (cycle Cycle) ValidTime(value string) bool {
	pattern, ok := cycleTimePatterns[cycle]
	if !ok {
		return false
	}
	return pattern.MatchString(value)
}

func (page Page) pathSegments(op string) ([]string, error) {
	if page.Start < 1 {
		return nil, validationError(op, "page.start", "must be greater than zero")
	}
	if page.End < 1 {
		return nil, validationError(op, "page.end", "must be greater than zero")
	}
	if page.End < page.Start {
		return nil, validationError(op, "page.end", "must be greater than or equal to page.start")
	}
	return []string{strconv.Itoa(page.Start), strconv.Itoa(page.End)}, nil
}

func requiredString(op string, field string, value string) error {
	if strings.TrimSpace(value) != "" {
		return nil
	}
	return validationError(op, field, "is required")
}

func validateCycle(op string, cycle Cycle) error {
	if strings.TrimSpace(string(cycle)) == "" {
		return validationError(op, "cycle", "is required")
	}
	if _, ok := cycleTimePatterns[cycle]; !ok {
		return validationError(op, "cycle", "is not supported")
	}
	return nil
}

func validateCycleTime(op string, field string, cycle Cycle, value string) error {
	if err := requiredString(op, field, value); err != nil {
		return err
	}
	if !cycle.ValidTime(value) {
		return oops.In("validation").
			With("op", op, "field", field, "cycle", string(cycle)).
			Errorf("ecos: %s must match ECOS %s time format", field, cycle)
	}
	return nil
}

func validationError(op string, field string, detail string) error {
	return oops.In("validation").
		With("op", op, "field", field).
		Errorf("ecos: %s %s", field, detail)
}
