package store_test

import (
	"path/filepath"
	"testing"

	"theatre39/internal/domain"
	"theatre39/internal/store"
)

func TestStoreRoundTrip(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "store.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	batch := domain.ReservationBatch{ID: "batch-1", Theatre: "Hall", Performance: "sat", Status: domain.BatchSubmitted, ItemIDs: []string{"item-1"}}
	if err := db.SaveBatch(batch); err != nil {
		t.Fatal(err)
	}
	if err := db.SaveItem(domain.ReservationItem{ID: "item-1", BatchID: batch.ID, Position: 1, Patron: "Ada", SeatCode: "B-10", Status: domain.ItemPending}); err != nil {
		t.Fatal(err)
	}
	if err := db.SaveHold(domain.SeatHold{ID: "hold-item-1", BatchID: batch.ID, ItemID: "item-1", SeatCode: "B-10", State: domain.HoldActive}); err != nil {
		t.Fatal(err)
	}
	got, err := db.GetBatch(batch.ID)
	if err != nil || got.Theatre != "Hall" {
		t.Fatalf("batch round trip failed: %+v %v", got, err)
	}
	items, err := db.ListItems(batch.ID)
	if err != nil || len(items) != 1 {
		t.Fatalf("item round trip failed: %+v %v", items, err)
	}
}

func TestStoreMissingRecord(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "store.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.GetBatch("missing"); err != store.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
