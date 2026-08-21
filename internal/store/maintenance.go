package store

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"go.etcd.io/bbolt"

	"theatre39/internal/domain"
)

type IntegrityReport struct {
	BatchCount       int
	ItemCount        int
	HoldCount        int
	ReviewCount      int
	ArchiveCount     int
	OrphanItems      []string
	OrphanHolds      []string
	MissingBatchRefs []string
}

func (d *DB) InspectIntegrity() (IntegrityReport, error) {
	report := IntegrityReport{}
	batches, err := queryAll(d, "batches", func(data []byte) (domain.ReservationBatch, error) {
		var value domain.ReservationBatch
		return value, decode(data, &value)
	})
	if err != nil {
		return report, err
	}
	report.BatchCount = len(batches)
	knownBatches := make(map[string]bool, len(batches))
	for _, batch := range batches {
		knownBatches[batch.ID] = true
	}
	items, err := queryAll(d, "items", func(data []byte) (domain.ReservationItem, error) {
		var value domain.ReservationItem
		return value, decode(data, &value)
	})
	if err != nil {
		return report, err
	}
	report.ItemCount = len(items)
	for _, item := range items {
		if !knownBatches[item.BatchID] {
			report.OrphanItems = append(report.OrphanItems, item.ID)
		}
	}
	holds, err := queryAll(d, "holds", func(data []byte) (domain.SeatHold, error) {
		var value domain.SeatHold
		return value, decode(data, &value)
	})
	if err != nil {
		return report, err
	}
	report.HoldCount = len(holds)
	knownItems := make(map[string]bool, len(items))
	for _, item := range items {
		knownItems[item.ID] = true
	}
	for _, hold := range holds {
		if !knownItems[hold.ItemID] {
			report.OrphanHolds = append(report.OrphanHolds, hold.ID)
		}
	}
	reviews, err := queryAll(d, "reviews", func(data []byte) (domain.ReviewRecord, error) {
		var value domain.ReviewRecord
		return value, decode(data, &value)
	})
	if err != nil {
		return report, err
	}
	report.ReviewCount = len(reviews)
	archives, err := queryAll(d, "archives", func(data []byte) (domain.ArchiveEntry, error) {
		var value domain.ArchiveEntry
		return value, decode(data, &value)
	})
	if err != nil {
		return report, err
	}
	report.ArchiveCount = len(archives)
	sort.Strings(report.OrphanItems)
	sort.Strings(report.OrphanHolds)
	return report, nil
}

func (d *DB) ReleaseHold(holdID, reason string) error {
	return d.db.Update(func(tx *bbolt.Tx) error {
		var hold domain.SeatHold
		if err := get(tx, "holds", holdID, &hold); err != nil {
			return err
		}
		if hold.State == domain.HoldReleased {
			return nil
		}
		hold.State = domain.HoldReleased
		if reason != "" {
			hold.ExpiresOn = reason
		}
		return put(tx, "holds", hold.ID, hold)
	})
}

func (d *DB) ReleaseBatchHolds(batchID, reason string) (int, error) {
	holds, err := d.ListHolds(batchID)
	if err != nil {
		return 0, err
	}
	changed := 0
	for _, hold := range holds {
		if hold.State != domain.HoldReleased {
			if err := d.ReleaseHold(hold.ID, reason); err != nil {
				return changed, err
			}
			changed++
		}
	}
	return changed, nil
}

func (d *DB) ExportBatch(batchID string) ([]byte, error) {
	batch, err := d.GetBatch(batchID)
	if err != nil {
		return nil, err
	}
	items, err := d.ListItems(batchID)
	if err != nil {
		return nil, err
	}
	reviews, err := d.ListReviews(batchID)
	if err != nil {
		return nil, err
	}
	holds, err := d.ListHolds(batchID)
	if err != nil {
		return nil, err
	}
	payload := struct {
		Batch   domain.ReservationBatch  `json:"batch"`
		Items   []domain.ReservationItem `json:"items"`
		Reviews []domain.ReviewRecord    `json:"reviews"`
		Holds   []domain.SeatHold        `json:"holds"`
	}{batch, items, reviews, holds}
	return json.MarshalIndent(payload, "", "  ")
}

func (d *DB) EventLog(batchID string) (string, error) {
	events, err := d.ListEvents(batchID)
	if err != nil {
		return "", err
	}
	lines := make([]string, 0, len(events))
	for _, event := range events {
		attributes := make([]string, 0, len(event.Attributes))
		for key, value := range event.Attributes {
			attributes = append(attributes, key+"="+value)
		}
		sort.Strings(attributes)
		lines = append(lines, strings.Join([]string{event.ID, event.Kind, event.Message, strings.Join(attributes, ",")}, " | "))
	}
	return strings.Join(lines, "\n"), nil
}

func (r IntegrityReport) Healthy() bool {
	return len(r.OrphanItems) == 0 && len(r.OrphanHolds) == 0 && len(r.MissingBatchRefs) == 0
}

func (r IntegrityReport) String() string {
	return fmt.Sprintf("batches=%d items=%d holds=%d reviews=%d archives=%d healthy=%t", r.BatchCount, r.ItemCount, r.HoldCount, r.ReviewCount, r.ArchiveCount, r.Healthy())
}
