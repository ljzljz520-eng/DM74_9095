package workflow39_test

import (
	"path/filepath"
	"testing"

	"theatre39/internal/domain"
	"theatre39/internal/query"
	"theatre39/internal/store"
	"theatre39/internal/workflow39"
)

func testCoordinator(t *testing.T) *workflow39.Coordinator {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "workflow.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return workflow39.New(db)
}

func TestWorkflowOne(t *testing.T) {
	coordinator := testCoordinator(t)
	report, err := coordinator.FullChain(domain.BatchInput{ID: "batch-one", Theatre: "Hall", Performance: "sat", CreatedBy: "desk", SubmittedNote: "front desk"}, []domain.ItemInput{{ID: "item-one", Patron: "Ada", SeatCode: "B-10", RequestedClass: "standard"}}, "Reviewer", "desk", "sat", "weekly archive")
	if err != nil {
		t.Fatal(err)
	}
	if report.ArchiveID == "" || report.Stage != "archive" {
		t.Fatalf("workflow one did not archive: %+v", report)
	}
}

func TestWorkflowTwo(t *testing.T) {
	coordinator := testCoordinator(t)
	if _, err := coordinator.RegisterAndSubmit(domain.BatchInput{ID: "batch-two", Theatre: "Hall", Performance: "sun", CreatedBy: "desk"}, []domain.ItemInput{{ID: "item-two", Patron: "Bea", SeatCode: "C-12", RequestedClass: "student"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := coordinator.ReviewAndNotify("batch-two", "Reviewer", "desk"); err != nil {
		t.Fatal(err)
	}
	page, err := coordinator.Query.FindBatches(query.BatchFilter{Theatre: "Hall"})
	if err != nil || page.Total != 1 {
		t.Fatalf("workflow two query failed: %+v %v", page, err)
	}
}

func TestWorkflowThree(t *testing.T) {
	coordinator := testCoordinator(t)
	items := []domain.ItemInput{{ID: "item-a", Patron: "Ada", SeatCode: "B-10", RequestedClass: "standard"}, {ID: "item-b", Patron: "Bo", SeatCode: "B-11", RequestedClass: "standard"}}
	report, err := coordinator.FullChain(domain.BatchInput{ID: "batch-three", Theatre: "Hall", Performance: "sun", CreatedBy: "desk"}, items, "Reviewer", "desk", "sun", "archive note")
	if err != nil || report.Completed != 2 {
		t.Fatalf("workflow three failed: %+v %v", report, err)
	}
}

func TestWorkflow39BusinessInvariant(t *testing.T) {
	coordinator := testCoordinator(t)
	items := []domain.ItemInput{{ID: "item-1", Patron: "P1", SeatCode: "B-10", RequestedClass: "standard"}, {ID: "item-2", Patron: "P2", SeatCode: "B-11", RequestedClass: "standard"}, {ID: "item-3", Patron: "P3", SeatCode: "B-12", RequestedClass: "standard"}, {ID: "item-4", Patron: "P4", SeatCode: "B-13", RequestedClass: "standard"}, {ID: "item-5", Patron: "P5", SeatCode: "A-01", RequestedClass: "standard"}, {ID: "item-6", Patron: "P6", SeatCode: "B-14", RequestedClass: "standard"}}
	if _, err := coordinator.RegisterAndSubmit(domain.BatchInput{ID: "partial-batch", Theatre: "Hall", Performance: "sat", CreatedBy: "desk"}, items); err != nil {
		t.Fatal(err)
	}
	report, err := coordinator.ReviewAndNotify("partial-batch", "Reviewer", "desk")
	if err != nil {
		t.Fatal(err)
	}
	if report.Notification != "partial" {
		t.Fatalf("mixed outcomes must publish partial notification, got %s", report.Notification)
	}
	stored, err := coordinator.Store.GetItem("item-5")
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status != domain.ItemRejected {
		t.Fatalf("failed item must remain rejected for follow-up, got %s", stored.Status)
	}
}
