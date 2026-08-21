package domain_test

import (
	"testing"

	"theatre39/internal/domain"
)

func TestValidateBatchAndItems(t *testing.T) {
	if err := domain.ValidateBatchInput(domain.BatchInput{ID: "b-1", Theatre: "Hall", Performance: "sat-19", CreatedBy: "desk"}); err != nil {
		t.Fatal(err)
	}
	items := []domain.ItemInput{{ID: "i-1", Patron: "Ada", SeatCode: "B-10", RequestedClass: "standard"}, {ID: "i-2", Patron: "Bo", SeatCode: "B-11", RequestedClass: "student"}}
	if err := domain.ValidateItemSet(items); err != nil {
		t.Fatal(err)
	}
	if got := domain.NormalizeItemInput(domain.ItemInput{ID: " i-3 ", Patron: " Cid ", SeatCode: "c-12", RequestedClass: " STANDARD "}); got.SeatCode != "C-12" || got.RequestedClass != "standard" {
		t.Fatalf("normalization failed: %+v", got)
	}
}

func TestRejectDuplicateSeat(t *testing.T) {
	err := domain.ValidateItemSet([]domain.ItemInput{{ID: "i-1", Patron: "Ada", SeatCode: "B-10", RequestedClass: "standard"}, {ID: "i-2", Patron: "Bo", SeatCode: "B-10", RequestedClass: "standard"}})
	if err == nil {
		t.Fatal("expected duplicate seat error")
	}
}

func TestSortItemsAndTerminalStates(t *testing.T) {
	items := domain.SortItems([]domain.ReservationItem{{ID: "b", Position: 2}, {ID: "a", Position: 1}})
	if items[0].ID != "a" {
		t.Fatalf("unexpected order: %+v", items)
	}
	if !domain.IsTerminalItem(domain.ItemCompleted) || domain.IsOpenBatch(domain.BatchArchived) {
		t.Fatal("state helpers returned wrong result")
	}
}
