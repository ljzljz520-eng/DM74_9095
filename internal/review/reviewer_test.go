package review_test

import (
	"path/filepath"
	"testing"

	"theatre39/internal/domain"
	"theatre39/internal/intake"
	"theatre39/internal/review"
	"theatre39/internal/store"
)

func TestReviewBatchOutcomes(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "review.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	intakeService := intake.NewService(db)
	_, err = intakeService.Register(domain.BatchInput{ID: "batch-1", Theatre: "Hall", Performance: "sat", CreatedBy: "desk"}, []domain.ItemInput{{ID: "item-1", Patron: "Ada", SeatCode: "B-10", RequestedClass: "standard"}, {ID: "item-2", Patron: "Bo", SeatCode: "A-01", RequestedClass: "standard"}})
	if err != nil {
		t.Fatal(err)
	}
	reviewer := review.NewService(db, review.DefaultPolicy())
	records, err := reviewer.ReviewBatch("batch-1", "Reviewer")
	if err != nil || len(records) != 2 {
		t.Fatalf("review failed: %v %v", records, err)
	}
	if records[0].Decision != domain.DecisionApprove || records[1].Decision != domain.DecisionReject {
		t.Fatalf("unexpected decisions: %+v", records)
	}
	batch, err := db.GetBatch("batch-1")
	if err != nil || batch.Status != domain.BatchPartial {
		t.Fatalf("unexpected batch state: %+v %v", batch, err)
	}
}

func TestReopenRejectedItems(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "review.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	service := intake.NewService(db)
	_, err = service.Register(domain.BatchInput{ID: "batch-2", Theatre: "Hall", Performance: "sun", CreatedBy: "desk"}, []domain.ItemInput{{ID: "item-1", Patron: "Ada", SeatCode: "A-01", RequestedClass: "standard"}})
	if err != nil {
		t.Fatal(err)
	}
	reviewer := review.NewService(db, review.DefaultPolicy())
	if _, err := reviewer.ReviewBatch("batch-2", "Reviewer"); err != nil {
		t.Fatal(err)
	}
	count, err := reviewer.ReopenRejected("batch-2", "Reviewer")
	if err != nil || count != 1 {
		t.Fatalf("reopen failed: %d %v", count, err)
	}
}
