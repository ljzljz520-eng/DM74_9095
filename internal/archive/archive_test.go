package archive_test

import (
	"path/filepath"
	"testing"

	"theatre39/internal/archive"
	"theatre39/internal/domain"
	"theatre39/internal/intake"
	"theatre39/internal/review"
	"theatre39/internal/store"
)

func TestArchiveAndRestore(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "archive.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	intakeService := intake.NewService(db)
	_, err = intakeService.Register(domain.BatchInput{ID: "batch-1", Theatre: "Hall", Performance: "sat", CreatedBy: "desk"}, []domain.ItemInput{{ID: "item-1", Patron: "Ada", SeatCode: "B-10", RequestedClass: "standard"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := review.NewService(db, review.DefaultPolicy()).ReviewBatch("batch-1", "Reviewer"); err != nil {
		t.Fatal(err)
	}
	service := archive.NewService(db, archive.DefaultRetentionPolicy())
	entry, err := service.Archive("batch-1", "Archivist", "sat", "weekly retention")
	if err != nil || entry.ID != "archive-batch-1" {
		t.Fatalf("archive failed: %+v %v", entry, err)
	}
	if err := service.Restore("batch-1", "Archivist"); err != nil {
		t.Fatal(err)
	}
	batch, err := db.GetBatch("batch-1")
	if err != nil || batch.Status != domain.BatchReviewed {
		t.Fatalf("restore failed: %+v %v", batch, err)
	}
}
