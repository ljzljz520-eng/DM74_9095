package archive

import (
	"fmt"
	"sort"
	"strings"

	"theatre39/internal/domain"
)

type LedgerRow struct {
	ArchiveID string
	BatchID   string
	Class     string
	Items     int
	Rejected  int
	Owner     string
}

func ToLedger(entries []domain.ArchiveEntry) []LedgerRow {
	rows := make([]LedgerRow, 0, len(entries))
	for _, entry := range entries {
		rows = append(rows, LedgerRow{ArchiveID: entry.ID, BatchID: entry.BatchID, Class: Classify(entry), Items: len(entry.ItemIDs), Rejected: len(entry.RejectedIDs), Owner: entry.ArchivedBy})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Class == rows[j].Class {
			return rows[i].BatchID < rows[j].BatchID
		}
		return rows[i].Class < rows[j].Class
	})
	return rows
}

func LedgerTotals(rows []LedgerRow) (int, int, int) {
	items := 0
	rejected := 0
	partial := 0
	for _, row := range rows {
		items += row.Items
		rejected += row.Rejected
		if row.Class == "partial" {
			partial++
		}
	}
	return items, rejected, partial
}

func RenderLedger(rows []LedgerRow) string {
	lines := []string{"archive_id,batch_id,class,items,rejected,owner"}
	for _, row := range rows {
		lines = append(lines, fmt.Sprintf("%s,%s,%s,%d,%d,%s", row.ArchiveID, row.BatchID, row.Class, row.Items, row.Rejected, row.Owner))
	}
	return strings.Join(lines, "\n")
}

func ValidateLedger(rows []LedgerRow) error {
	seen := make(map[string]bool, len(rows))
	for _, row := range rows {
		if row.ArchiveID == "" || row.BatchID == "" {
			return fmt.Errorf("ledger identity is missing")
		}
		if seen[row.ArchiveID] {
			return fmt.Errorf("ledger archive %s is duplicated", row.ArchiveID)
		}
		if row.Items < 1 {
			return fmt.Errorf("ledger archive %s has no items", row.ArchiveID)
		}
		if row.Rejected < 0 || row.Rejected > row.Items {
			return fmt.Errorf("ledger archive %s has invalid rejected count", row.ArchiveID)
		}
		seen[row.ArchiveID] = true
	}
	return nil
}

func FilterLedger(rows []LedgerRow, class, owner string) []LedgerRow {
	filtered := make([]LedgerRow, 0, len(rows))
	for _, row := range rows {
		if class != "" && row.Class != class {
			continue
		}
		if owner != "" && !strings.EqualFold(row.Owner, owner) {
			continue
		}
		filtered = append(filtered, row)
	}
	return filtered
}

func LedgerByBatch(rows []LedgerRow) map[string]LedgerRow {
	result := make(map[string]LedgerRow, len(rows))
	for _, row := range rows {
		result[row.BatchID] = row
	}
	return result
}

func (s *Service) Ledger(class, owner string) (string, error) {
	index, err := s.BuildIndex()
	if err != nil {
		return "", err
	}
	rows := FilterLedger(ToLedger(index.Entries), class, owner)
	if err := ValidateLedger(rows); err != nil {
		return "", err
	}
	items, rejected, partial := LedgerTotals(rows)
	return fmt.Sprintf("%s\ntotals items=%d rejected=%d partial=%d", RenderLedger(rows), items, rejected, partial), nil
}

func LedgerClass(row LedgerRow) string {
	if row.Rejected == 0 {
		return "complete"
	}
	return row.Class
}
