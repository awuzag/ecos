package cli

import (
	"bytes"
	"strings"
	"text/tabwriter"
	"unicode/utf8"

	"github.com/samber/oops"
)

func tableRows[T any](headers []string, rows []T, values func(T) []string) string {
	var buffer bytes.Buffer
	writer := tabwriter.NewWriter(&buffer, 0, 0, 2, ' ', 0)
	writeLine(writer, headers)
	for _, row := range rows {
		writeLine(writer, values(row))
	}
	if err := writer.Flush(); err != nil {
		return oops.In("cli").Wrap(err).Error()
	}
	return buffer.String()
}

func writeLine(buffer *tabwriter.Writer, values []string) {
	for index, value := range values {
		if index > 0 {
			_, _ = buffer.Write([]byte("\t"))
		}
		_, _ = buffer.Write([]byte(sanitizeCell(value)))
	}
	_, _ = buffer.Write([]byte("\n"))
}

func sanitizeCell(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func compactText(value string) string {
	const maxRunes = 80
	cleaned := sanitizeCell(value)
	if utf8.RuneCountInString(cleaned) <= maxRunes {
		return cleaned
	}
	runes := []rune(cleaned)
	return string(runes[:maxRunes-3]) + "..."
}
