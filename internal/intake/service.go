package intake

import (
	"fmt"
	"strings"

	"theatre39/internal/domain"
	"theatre39/internal/store"
)

type Service struct {
	store *store.DB
}

func NewService(db *store.DB) *Service {
	return &Service{store: db}
}

func (s *Service) Register(input domain.BatchInput, inputs []domain.ItemInput) (domain.ReservationBatch, error) {
	if err := domain.ValidateBatchInput(input); err != nil {
		return domain.ReservationBatch{}, err
	}
	if err := domain.ValidateItemSet(inputs); err != nil {
		return domain.ReservationBatch{}, err
	}
	if existing, err := s.store.GetBatch(input.ID); err == nil && existing.ID != "" {
		return domain.ReservationBatch{}, fmt.Errorf("batch %s already exists", input.ID)
	}
	items := make([]domain.ReservationItem, 0, len(inputs))
	itemIDs := make([]string, 0, len(inputs))
	for position, raw := range inputs {
		item := domain.NormalizeItemInput(raw)
		reservation := domain.ReservationItem{
			ID:             item.ID,
			BatchID:        input.ID,
			Position:       position + 1,
			Patron:         item.Patron,
			SeatCode:       item.SeatCode,
			RequestedClass: item.RequestedClass,
			Status:         domain.ItemPending,
		}
		items = append(items, reservation)
		itemIDs = append(itemIDs, reservation.ID)
	}
	batch := domain.ReservationBatch{
		ID:            strings.TrimSpace(input.ID),
		Theatre:       strings.TrimSpace(input.Theatre),
		Performance:   strings.TrimSpace(input.Performance),
		CreatedBy:     strings.TrimSpace(input.CreatedBy),
		Status:        domain.BatchSubmitted,
		ItemIDs:       itemIDs,
		SubmittedNote: strings.TrimSpace(input.SubmittedNote),
		Version:       1,
	}
	if err := s.store.SaveBatch(batch); err != nil {
		return domain.ReservationBatch{}, err
	}
	for _, item := range items {
		if err := s.store.SaveItem(item); err != nil {
			return domain.ReservationBatch{}, fmt.Errorf("save item %s: %w", item.ID, err)
		}
		hold := domain.SeatHold{
			ID:        "hold-" + item.ID,
			BatchID:   batch.ID,
			ItemID:    item.ID,
			SeatCode:  item.SeatCode,
			State:     domain.HoldActive,
			ExpiresOn: batch.Performance,
		}
		if err := s.store.SaveHold(hold); err != nil {
			return domain.ReservationBatch{}, fmt.Errorf("save hold %s: %w", hold.ID, err)
		}
	}
	if err := s.store.RecordEvent(domain.BatchEvent{
		ID:      "event-register-" + batch.ID,
		BatchID: batch.ID,
		Kind:    "registered",
		Message: fmt.Sprintf("batch %s registered with %d items", batch.ID, len(items)),
	}); err != nil {
		return domain.ReservationBatch{}, err
	}
	return batch, nil
}

func (s *Service) AddItem(batchID string, input domain.ItemInput) (domain.ReservationItem, error) {
	batch, err := s.store.GetBatch(batchID)
	if err != nil {
		return domain.ReservationItem{}, err
	}
	if batch.Status != domain.BatchDraft && batch.Status != domain.BatchSubmitted {
		return domain.ReservationItem{}, fmt.Errorf("batch %s is not editable", batchID)
	}
	if err := domain.ValidateItemInput(input); err != nil {
		return domain.ReservationItem{}, err
	}
	item := domain.NormalizeItemInput(input)
	for _, existing := range batch.ItemIDs {
		if existing == item.ID {
			return domain.ReservationItem{}, fmt.Errorf("item %s already exists", item.ID)
		}
	}
	items, err := s.store.ListItems(batchID)
	if err != nil {
		return domain.ReservationItem{}, err
	}
	for _, existing := range items {
		if existing.SeatCode == item.SeatCode {
			return domain.ReservationItem{}, fmt.Errorf("seat %s is already held", item.SeatCode)
		}
	}
	reservation := domain.ReservationItem{
		ID:             item.ID,
		BatchID:        batchID,
		Position:       len(items) + 1,
		Patron:         item.Patron,
		SeatCode:       item.SeatCode,
		RequestedClass: item.RequestedClass,
		Status:         domain.ItemPending,
	}
	if err := s.store.SaveItem(reservation); err != nil {
		return domain.ReservationItem{}, err
	}
	batch.ItemIDs = append(batch.ItemIDs, reservation.ID)
	batch.Version++
	if err := s.store.SaveBatch(batch); err != nil {
		return domain.ReservationItem{}, err
	}
	if err := s.store.SaveHold(domain.SeatHold{ID: "hold-" + item.ID, BatchID: batchID, ItemID: item.ID, SeatCode: item.SeatCode, State: domain.HoldActive, ExpiresOn: batch.Performance}); err != nil {
		return domain.ReservationItem{}, err
	}
	return reservation, nil
}

func (s *Service) Submit(batchID, note string) error {
	batch, err := s.store.GetBatch(batchID)
	if err != nil {
		return err
	}
	if len(batch.ItemIDs) == 0 {
		return fmt.Errorf("cannot submit empty batch")
	}
	if batch.Status != domain.BatchDraft && batch.Status != domain.BatchSubmitted {
		return fmt.Errorf("batch %s is already reviewed", batchID)
	}
	batch.Status = domain.BatchSubmitted
	batch.SubmittedNote = strings.TrimSpace(note)
	batch.Version++
	if err := s.store.SaveBatch(batch); err != nil {
		return err
	}
	return s.store.RecordEvent(domain.BatchEvent{ID: "event-submit-" + batchID, BatchID: batchID, Kind: "submitted", Message: "batch submitted for review"})
}
