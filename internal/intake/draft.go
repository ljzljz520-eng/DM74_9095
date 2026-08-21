package intake

import (
	"fmt"

	"theatre39/internal/domain"
)

func (s *Service) StartDraft(input domain.BatchInput) (domain.ReservationBatch, error) {
	if err := domain.ValidateBatchInput(input); err != nil {
		return domain.ReservationBatch{}, err
	}
	if _, err := s.store.GetBatch(input.ID); err == nil {
		return domain.ReservationBatch{}, fmt.Errorf("batch %s already exists", input.ID)
	}
	batch := domain.ReservationBatch{ID: input.ID, Theatre: input.Theatre, Performance: input.Performance, CreatedBy: input.CreatedBy, Status: domain.BatchDraft, SubmittedNote: input.SubmittedNote, Version: 1}
	if err := s.store.SaveBatch(batch); err != nil {
		return domain.ReservationBatch{}, err
	}
	if err := s.store.RecordEvent(domain.BatchEvent{ID: "event-draft-" + batch.ID, BatchID: batch.ID, Kind: "drafted", Message: "draft opened"}); err != nil {
		return domain.ReservationBatch{}, err
	}
	return batch, nil
}

func (s *Service) RemoveItem(batchID, itemID string) error {
	batch, err := s.store.GetBatch(batchID)
	if err != nil {
		return err
	}
	if batch.Status != domain.BatchDraft {
		return fmt.Errorf("only draft batches can remove items")
	}
	items, err := s.store.ListItems(batchID)
	if err != nil {
		return err
	}
	found := false
	remaining := make([]string, 0, len(batch.ItemIDs))
	for _, id := range batch.ItemIDs {
		if id == itemID {
			found = true
			continue
		}
		remaining = append(remaining, id)
	}
	if !found {
		return fmt.Errorf("item %s is not in batch", itemID)
	}
	for _, item := range items {
		if item.ID == itemID {
			if err := s.store.ReleaseHold("hold-"+item.ID, "removed from draft"); err != nil {
				return err
			}
		}
	}
	batch.ItemIDs = remaining
	batch.Version++
	return s.store.SaveBatch(batch)
}

func (s *Service) ReplaceItem(batchID string, input domain.ItemInput) (domain.ReservationItem, error) {
	if err := domain.ValidateItemInput(input); err != nil {
		return domain.ReservationItem{}, err
	}
	batch, err := s.store.GetBatch(batchID)
	if err != nil {
		return domain.ReservationItem{}, err
	}
	if batch.Status != domain.BatchDraft {
		return domain.ReservationItem{}, fmt.Errorf("only draft batches can replace items")
	}
	old, err := s.store.GetItem(input.ID)
	if err != nil {
		return domain.ReservationItem{}, err
	}
	if old.BatchID != batchID {
		return domain.ReservationItem{}, fmt.Errorf("item belongs to a different batch")
	}
	normalized := domain.NormalizeItemInput(input)
	replacement := domain.ReservationItem{ID: normalized.ID, BatchID: batchID, Position: old.Position, Patron: normalized.Patron, SeatCode: normalized.SeatCode, RequestedClass: normalized.RequestedClass, Status: domain.ItemPending}
	if err := s.store.SaveItem(replacement); err != nil {
		return domain.ReservationItem{}, err
	}
	if err := s.store.ReleaseHold("hold-"+old.ID, "replaced in draft"); err != nil {
		return domain.ReservationItem{}, err
	}
	if err := s.store.SaveHold(domain.SeatHold{ID: "hold-" + replacement.ID, BatchID: batchID, ItemID: replacement.ID, SeatCode: replacement.SeatCode, State: domain.HoldActive, ExpiresOn: batch.Performance}); err != nil {
		return domain.ReservationItem{}, err
	}
	return replacement, nil
}
