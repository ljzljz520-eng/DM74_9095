package query

import (
	"fmt"
	"sort"
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

func (s *Service) FindBatches(filter BatchFilter) (domain.Page[domain.ReservationBatch], error) {
	filter = filter.Normalize()
	batches, err := s.store.FindBatches(filter.Theatre, filter.Performance, filter.Status)
	if err != nil {
		return domain.Page[domain.ReservationBatch]{}, err
	}
	if filter.Patron == "" && filter.SeatCode == "" {
		sort.Slice(batches, func(i, j int) bool { return batches[i].ID < batches[j].ID })
		return domain.Page[domain.ReservationBatch]{Items: batches, Total: len(batches)}, nil
	}
	matched := make([]domain.ReservationBatch, 0, len(batches))
	for _, batch := range batches {
		items, listErr := s.store.ListItems(batch.ID)
		if listErr != nil {
			return domain.Page[domain.ReservationBatch]{}, listErr
		}
		for _, item := range items {
			if matchesItem(item, filter) {
				matched = append(matched, batch)
				break
			}
		}
	}
	sort.Slice(matched, func(i, j int) bool { return matched[i].ID < matched[j].ID })
	return domain.Page[domain.ReservationBatch]{Items: matched, Total: len(matched)}, nil
}

func (s *Service) GetBatchDetail(batchID string) (domain.ReservationBatch, []domain.ReservationItem, []domain.ReviewRecord, error) {
	batch, err := s.store.GetBatch(batchID)
	if err != nil {
		return domain.ReservationBatch{}, nil, nil, err
	}
	items, err := s.store.ListItems(batchID)
	if err != nil {
		return domain.ReservationBatch{}, nil, nil, err
	}
	reviews, err := s.store.ListReviews(batchID)
	if err != nil {
		return domain.ReservationBatch{}, nil, nil, err
	}
	return batch, domain.SortItems(items), reviews, nil
}

func (s *Service) GetAvailability(batchID string) (map[string]string, error) {
	holds, err := s.store.ListHolds(batchID)
	if err != nil {
		return nil, err
	}
	availability := make(map[string]string, len(holds))
	for _, hold := range holds {
		availability[hold.SeatCode] = hold.State
	}
	return availability, nil
}

func (s *Service) Summarize(batchID string) (domain.BatchSummary, error) {
	items, err := s.store.ListItems(batchID)
	if err != nil {
		return domain.BatchSummary{}, err
	}
	summary := summarizeItems(items)
	for _, item := range items {
		if item.Status == domain.ItemRejected {
			summary.FailedItemIDs = append(summary.FailedItemIDs, item.ID)
		}
	}
	return summary, nil
}

func (s *Service) RenderStatus(batchID string) (string, error) {
	batch, items, _, err := s.GetBatchDetail(batchID)
	if err != nil {
		return "", err
	}
	summary := summarizeItems(items)
	return fmt.Sprintf("%s/%s: %s (%d approved, %d rejected, %d pending)", batch.Theatre, batch.Performance, strings.ToUpper(summary.Kind), summary.Approved, summary.Rejected, summary.Pending), nil
}
