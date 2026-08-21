package intake_test

import (
	"path/filepath"
	"testing"

	"theatre39/internal/domain"
	"theatre39/internal/intake"
	"theatre39/internal/store"
)

func TestRegisterBatch(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "intake.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	service := intake.NewService(db)
	batch, err := service.Register(domain.BatchInput{ID: "batch-1", Theatre: "Hall", Performance: "sat", CreatedBy: "desk"}, []domain.ItemInput{{ID: "item-1", Patron: "Ada", SeatCode: "B-10", RequestedClass: "standard"}})
	if err != nil {
		t.Fatal(err)
	}
	if batch.Status != domain.BatchSubmitted || len(batch.ItemIDs) != 1 {
		t.Fatalf("unexpected batch: %+v", batch)
	}
	if err := service.Submit(batch.ID, "ready"); err != nil {
		t.Fatal(err)
	}
}

func TestParseRowsReportsInvalidRows(t *testing.T) {
	items, issues := intake.ParseRows([][]string{{"i-1", "Ada", "B-10", "standard"}, {"i-2", "Bo"}})
	if len(items) != 1 || len(issues) != 1 {
		t.Fatalf("unexpected parse result: items=%v issues=%v", items, issues)
	}
	if intake.FormatIssues(issues) == "" {
		t.Fatal("expected formatted issue")
	}
}
