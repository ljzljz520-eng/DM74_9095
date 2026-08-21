package workflow39

import (
	"fmt"
	"strings"

	"theatre39/internal/domain"
)

type Report struct {
	BatchID      string
	Stage        string
	Status       string
	Approved     int
	Rejected     int
	Completed    int
	Pending      int
	Notification string
	ArchiveID    string
	Warnings     []string
}

func NewReport(batchID string) Report {
	return Report{BatchID: batchID, Stage: "registration", Status: "started"}
}

func (r *Report) ApplySummary(summary domain.BatchSummary) {
	r.Approved = summary.Approved
	r.Rejected = summary.Rejected
	r.Pending = summary.Pending
	r.Status = summary.Kind
	if summary.Rejected > 0 && summary.Approved > 0 {
		r.Warnings = append(r.Warnings, "batch has mixed review outcomes")
	}
}

func (r Report) String() string {
	return fmt.Sprintf("%s stage=%s status=%s approved=%d rejected=%d completed=%d pending=%d notification=%s archive=%s warnings=%s", r.BatchID, r.Stage, r.Status, r.Approved, r.Rejected, r.Completed, r.Pending, r.Notification, r.ArchiveID, strings.Join(r.Warnings, "|"))
}

func CountCompleted(items []domain.ReservationItem) int {
	count := 0
	for _, item := range items {
		if item.Status == domain.ItemCompleted {
			count++
		}
	}
	return count
}
