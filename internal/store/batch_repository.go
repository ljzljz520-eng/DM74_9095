package store

import (
	"fmt"

	"go.etcd.io/bbolt"

	"theatre39/internal/domain"
)

func (d *DB) SaveBatch(batch domain.ReservationBatch) error {
	if batch.ID == "" {
		return fmt.Errorf("batch id is required")
	}
	return d.db.Update(func(tx *bbolt.Tx) error {
		return put(tx, "batches", batch.ID, batch)
	})
}

func (d *DB) GetBatch(id string) (domain.ReservationBatch, error) {
	var batch domain.ReservationBatch
	err := d.db.View(func(tx *bbolt.Tx) error {
		return get(tx, "batches", id, &batch)
	})
	return batch, err
}

func (d *DB) SaveItem(item domain.ReservationItem) error {
	if item.ID == "" || item.BatchID == "" {
		return fmt.Errorf("item and batch ids are required")
	}
	return d.db.Update(func(tx *bbolt.Tx) error {
		return put(tx, "items", item.ID, item)
	})
}

func (d *DB) GetItem(id string) (domain.ReservationItem, error) {
	var item domain.ReservationItem
	err := d.db.View(func(tx *bbolt.Tx) error {
		return get(tx, "items", id, &item)
	})
	return item, err
}

func (d *DB) ListItems(batchID string) ([]domain.ReservationItem, error) {
	return listInBatch(d, "items", batchID, func(data []byte) (domain.ReservationItem, error) {
		var item domain.ReservationItem
		return item, decode(data, &item)
	}, func(item domain.ReservationItem) bool { return item.BatchID == batchID })
}

func listInBatch[T any](d *DB, bucket, batchID string, decodeTarget func([]byte) (T, error), belongs func(T) bool) ([]T, error) {
	var all []T
	err := d.db.View(func(tx *bbolt.Tx) error {
		items, err := list(tx, bucket, decodeTarget)
		if err != nil {
			return err
		}
		for _, item := range items {
			if belongs(item) {
				all = append(all, item)
			}
		}
		return nil
	})
	return all, err
}

func (d *DB) SaveHold(hold domain.SeatHold) error {
	if hold.ID == "" || hold.SeatCode == "" {
		return fmt.Errorf("hold id and seat are required")
	}
	return d.db.Update(func(tx *bbolt.Tx) error {
		return put(tx, "holds", hold.ID, hold)
	})
}

func (d *DB) GetHold(id string) (domain.SeatHold, error) {
	var hold domain.SeatHold
	err := d.db.View(func(tx *bbolt.Tx) error {
		return get(tx, "holds", id, &hold)
	})
	return hold, err
}

func (d *DB) ListHolds(batchID string) ([]domain.SeatHold, error) {
	return listInBatch(d, "holds", batchID, func(data []byte) (domain.SeatHold, error) {
		var hold domain.SeatHold
		return hold, decode(data, &hold)
	}, func(hold domain.SeatHold) bool { return hold.BatchID == batchID })
}

func (d *DB) ListItemsByStatus(batchID, status string) ([]domain.ReservationItem, error) {
	items, err := d.ListItems(batchID)
	if err != nil {
		return nil, err
	}
	filtered := make([]domain.ReservationItem, 0, len(items))
	for _, item := range items {
		if item.Status == status {
			filtered = append(filtered, item)
		}
	}
	return filtered, nil
}

func (d *DB) UpdateItemStatus(itemID, status, reason, reviewer string) error {
	return d.db.Update(func(tx *bbolt.Tx) error {
		var item domain.ReservationItem
		if err := get(tx, "items", itemID, &item); err != nil {
			return err
		}
		item.Status = status
		item.Reason = reason
		item.ReviewedBy = reviewer
		return put(tx, "items", item.ID, item)
	})
}

func (d *DB) UpdateBatchStatus(batchID, status, reviewer, archiveID string) error {
	return d.db.Update(func(tx *bbolt.Tx) error {
		var batch domain.ReservationBatch
		if err := get(tx, "batches", batchID, &batch); err != nil {
			return err
		}
		batch.Status = status
		batch.ReviewedBy = reviewer
		batch.ArchiveID = archiveID
		batch.Version++
		return put(tx, "batches", batch.ID, batch)
	})
}
