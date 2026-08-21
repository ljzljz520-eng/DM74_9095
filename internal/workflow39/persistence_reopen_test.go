package workflow39_test

import (
	"path/filepath"
	"testing"

	"theatre39/internal/domain"
	"theatre39/internal/store"
	"theatre39/internal/workflow39"
)

func TestPersistenceSurvivesReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "persistent.db")
	db, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	coordinator := workflow39.New(db)
	if _, err := coordinator.RegisterAndSubmit(domain.BatchInput{ID: "persistent-batch", Theatre: "Hall", Performance: "fri", CreatedBy: "desk"}, []domain.ItemInput{{ID: "persistent-item", Patron: "Ada", SeatCode: "C-20", RequestedClass: "standard"}}); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	batch, err := reopened.GetBatch("persistent-batch")
	if err != nil || batch.ID != "persistent-batch" {
		t.Fatalf("batch did not survive reopen: %+v %v", batch, err)
	}
	items, err := reopened.ListItems("persistent-batch")
	if err != nil || len(items) != 1 || items[0].SeatCode != "C-20" {
		t.Fatalf("items did not survive reopen: %+v %v", items, err)
	}
}
