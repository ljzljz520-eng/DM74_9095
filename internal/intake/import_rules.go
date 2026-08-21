package intake

import (
	"fmt"
	"strings"

	"theatre39/internal/domain"
)

type RowIssue struct {
	Row     int
	Field   string
	Message string
}

func ParseRows(rows [][]string) ([]domain.ItemInput, []RowIssue) {
	items := make([]domain.ItemInput, 0, len(rows))
	issues := make([]RowIssue, 0)
	for index, row := range rows {
		if len(row) < 4 {
			issues = append(issues, RowIssue{Row: index + 1, Field: "columns", Message: "four columns are required"})
			continue
		}
		item := domain.ItemInput{ID: row[0], Patron: row[1], SeatCode: row[2], RequestedClass: row[3]}
		if err := domain.ValidateItemInput(item); err != nil {
			issues = append(issues, RowIssue{Row: index + 1, Field: "item", Message: err.Error()})
			continue
		}
		items = append(items, domain.NormalizeItemInput(item))
	}
	if len(items) == 0 && len(issues) == 0 {
		issues = append(issues, RowIssue{Row: 0, Field: "rows", Message: "no rows provided"})
	}
	return items, issues
}

func FormatIssues(issues []RowIssue) string {
	if len(issues) == 0 {
		return ""
	}
	parts := make([]string, 0, len(issues))
	for _, issue := range issues {
		parts = append(parts, fmt.Sprintf("row %d %s: %s", issue.Row, issue.Field, issue.Message))
	}
	return strings.Join(parts, "; ")
}

func MergeRows(base []domain.ItemInput, additions [][]string) ([]domain.ItemInput, []RowIssue) {
	parsed, issues := ParseRows(additions)
	merged := append([]domain.ItemInput(nil), base...)
	seen := make(map[string]bool, len(base))
	for _, item := range base {
		seen[item.ID] = true
	}
	for _, item := range parsed {
		if seen[item.ID] {
			issues = append(issues, RowIssue{Row: 0, Field: "id", Message: "duplicate item " + item.ID})
			continue
		}
		seen[item.ID] = true
		merged = append(merged, item)
	}
	return merged, issues
}
