package archive

import (
	"fmt"
	"strings"

	"theatre39/internal/domain"
)

type RetentionPolicy struct {
	AllowedStatuses map[string]bool
	RequiredNote    bool
}

func DefaultRetentionPolicy() RetentionPolicy {
	return RetentionPolicy{AllowedStatuses: map[string]bool{domain.BatchReviewed: true, domain.BatchPartial: true}, RequiredNote: true}
}

func (p RetentionPolicy) Validate(batch domain.ReservationBatch, note string) error {
	if !p.AllowedStatuses[batch.Status] {
		return fmt.Errorf("batch status %s cannot be archived", batch.Status)
	}
	if p.RequiredNote && strings.TrimSpace(note) == "" {
		return fmt.Errorf("archive note is required")
	}
	return nil
}

func Classify(entry domain.ArchiveEntry) string {
	if len(entry.RejectedIDs) == 0 {
		return "complete"
	}
	if len(entry.ItemIDs) == len(entry.RejectedIDs) {
		return "rejected"
	}
	return "partial"
}

func BuildSummary(entry domain.ArchiveEntry) string {
	return fmt.Sprintf("%s archive %s: %d items, %d rejected", Classify(entry), entry.ID, len(entry.ItemIDs), len(entry.RejectedIDs))
}
