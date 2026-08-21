package query

import (
	"strings"

	"theatre39/internal/domain"
)

type BatchFilter struct {
	Theatre     string
	Performance string
	Status      string
	Patron      string
	SeatCode    string
}

func (f BatchFilter) Normalize() BatchFilter {
	f.Theatre = strings.TrimSpace(f.Theatre)
	f.Performance = strings.TrimSpace(f.Performance)
	f.Status = strings.ToLower(strings.TrimSpace(f.Status))
	f.Patron = strings.TrimSpace(f.Patron)
	f.SeatCode = strings.ToUpper(strings.TrimSpace(f.SeatCode))
	return f
}

func matchesItem(item domain.ReservationItem, filter BatchFilter) bool {
	if filter.Patron != "" && !strings.Contains(strings.ToLower(item.Patron), strings.ToLower(filter.Patron)) {
		return false
	}
	if filter.SeatCode != "" && item.SeatCode != filter.SeatCode {
		return false
	}
	return true
}

func summarizeItems(items []domain.ReservationItem) domain.BatchSummary {
	summary := domain.BatchSummary{Total: len(items)}
	for _, item := range items {
		switch item.Status {
		case domain.ItemApproved:
			summary.Approved++
		case domain.ItemRejected:
			summary.Rejected++
		case domain.ItemPending:
			summary.Pending++
		}
	}
	summary.Kind = "empty"
	if summary.Total > 0 && summary.Rejected == 0 && summary.Pending == 0 {
		summary.Kind = "complete"
	}
	if summary.Rejected > 0 && summary.Approved > 0 {
		summary.Kind = "partial"
	}
	if summary.Rejected == summary.Total && summary.Total > 0 {
		summary.Kind = "rejected"
	}
	return summary
}
