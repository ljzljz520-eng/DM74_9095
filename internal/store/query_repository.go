package store

import (
	"fmt"
	"strings"

	"go.etcd.io/bbolt"

	"theatre39/internal/domain"
)

func (d *DB) SaveReview(review domain.ReviewRecord) error {
	if review.ID == "" || review.ItemID == "" {
		return fmt.Errorf("review id and item id are required")
	}
	return d.db.Update(func(tx *bbolt.Tx) error {
		return put(tx, "reviews", review.ID, review)
	})
}

func (d *DB) ListReviews(batchID string) ([]domain.ReviewRecord, error) {
	return listInBatch(d, "reviews", batchID, func(data []byte) (domain.ReviewRecord, error) {
		var review domain.ReviewRecord
		return review, decode(data, &review)
	}, func(review domain.ReviewRecord) bool { return review.BatchID == batchID })
}

func (d *DB) SaveArchive(entry domain.ArchiveEntry) error {
	if entry.ID == "" || entry.BatchID == "" {
		return fmt.Errorf("archive id and batch id are required")
	}
	return d.db.Update(func(tx *bbolt.Tx) error {
		return put(tx, "archives", entry.ID, entry)
	})
}

func (d *DB) GetArchive(id string) (domain.ArchiveEntry, error) {
	var entry domain.ArchiveEntry
	err := d.db.View(func(tx *bbolt.Tx) error {
		return get(tx, "archives", id, &entry)
	})
	return entry, err
}

func (d *DB) ListArchives() ([]domain.ArchiveEntry, error) {
	return queryAll(d, "archives", func(data []byte) (domain.ArchiveEntry, error) {
		var entry domain.ArchiveEntry
		return entry, decode(data, &entry)
	})
}

func (d *DB) ListEvents(batchID string) ([]domain.BatchEvent, error) {
	return listInBatch(d, "events", batchID, func(data []byte) (domain.BatchEvent, error) {
		var event domain.BatchEvent
		return event, decode(data, &event)
	}, func(event domain.BatchEvent) bool { return event.BatchID == batchID })
}

func queryAll[T any](d *DB, bucket string, decodeTarget func([]byte) (T, error)) ([]T, error) {
	var records []T
	err := d.db.View(func(tx *bbolt.Tx) error {
		items, err := list(tx, bucket, decodeTarget)
		if err != nil {
			return err
		}
		records = append(records, items...)
		return nil
	})
	return records, err
}

func (d *DB) FindBatches(theatre, performance, status string) ([]domain.ReservationBatch, error) {
	batches, err := queryAll(d, "batches", func(data []byte) (domain.ReservationBatch, error) {
		var batch domain.ReservationBatch
		return batch, decode(data, &batch)
	})
	if err != nil {
		return nil, err
	}
	filtered := make([]domain.ReservationBatch, 0, len(batches))
	for _, batch := range batches {
		if theatre != "" && !strings.EqualFold(batch.Theatre, theatre) {
			continue
		}
		if performance != "" && batch.Performance != performance {
			continue
		}
		if status != "" && batch.Status != status {
			continue
		}
		filtered = append(filtered, batch)
	}
	return filtered, nil
}
