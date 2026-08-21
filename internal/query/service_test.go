package query_test

import (
	"path/filepath"
	"testing"

	"theatre39/internal/domain"
	"theatre39/internal/intake"
	"theatre39/internal/query"
	"theatre39/internal/review"
	"theatre39/internal/store"
)

func TestQueryBatchAndAvailability(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "query.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	intakeService := intake.NewService(db)
	_, err = intakeService.Register(domain.BatchInput{ID: "batch-1", Theatre: "Hall", Performance: "sat", CreatedBy: "desk"}, []domain.ItemInput{{ID: "item-1", Patron: "Ada", SeatCode: "B-10", RequestedClass: "standard"}})
	if err != nil {
		t.Fatal(err)
	}
	reviewer := review.NewService(db, review.DefaultPolicy())
	if _, err := reviewer.ReviewBatch("batch-1", "Reviewer"); err != nil {
		t.Fatal(err)
	}
	service := query.NewService(db)
	page, err := service.FindBatches(query.BatchFilter{Theatre: "hall", Patron: "ada"})
	if err != nil || page.Total != 1 {
		t.Fatalf("filter failed: %+v %v", page, err)
	}
	availability, err := service.GetAvailability("batch-1")
	if err != nil || availability["B-10"] != domain.HoldActive {
		t.Fatalf("availability failed: %v %v", availability, err)
	}
	status, err := service.RenderStatus("batch-1")
	if err != nil || status == "" {
		t.Fatalf("render failed: %q %v", status, err)
	}
}
