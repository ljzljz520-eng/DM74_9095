package review

import (
	"fmt"
	"sort"

	"theatre39/internal/domain"
	"theatre39/internal/store"
)

type Service struct {
	store  *store.DB
	policy Policy
}

func NewService(db *store.DB, policy Policy) *Service {
	return &Service{store: db, policy: policy}
}

func (s *Service) ReviewBatch(batchID, reviewer string) ([]domain.ReviewRecord, error) {
	if err := ValidateReviewer(reviewer); err != nil {
		return nil, err
	}
	batch, err := s.store.GetBatch(batchID)
	if err != nil {
		return nil, err
	}
	if batch.Status != domain.BatchSubmitted && batch.Status != domain.BatchDraft {
		return nil, fmt.Errorf("batch %s is not awaiting review", batchID)
	}
	items, err := s.store.ListItems(batchID)
	if err != nil {
		return nil, err
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Position < items[j].Position })
	records := make([]domain.ReviewRecord, 0, len(items))
	approved := 0
	for sequence, item := range items {
		decision, reason := s.policy.Evaluate(item)
		if decision == domain.DecisionApprove {
			approved++
			if err := s.store.UpdateItemStatus(item.ID, domain.ItemApproved, reason, reviewer); err != nil {
				return nil, err
			}
		} else {
			if err := s.store.UpdateItemStatus(item.ID, domain.ItemRejected, reason, reviewer); err != nil {
				return nil, err
			}
		}
		record := domain.ReviewRecord{ID: fmt.Sprintf("review-%s-%03d", batchID, sequence+1), BatchID: batchID, ItemID: item.ID, Reviewer: reviewer, Decision: decision, Reason: reason, Sequence: sequence + 1}
		if err := s.store.SaveReview(record); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	newStatus := domain.BatchReviewed
	if approved == 0 {
		newStatus = domain.BatchPartial
	}
	if approved > 0 && approved < len(items) {
		newStatus = domain.BatchPartial
	}
	if err := s.store.UpdateBatchStatus(batchID, newStatus, reviewer, ""); err != nil {
		return nil, err
	}
	if err := s.store.RecordEvent(domain.BatchEvent{ID: "event-review-" + batchID, BatchID: batchID, Kind: "reviewed", Message: fmt.Sprintf("reviewed %d items", len(records)), Attributes: map[string]string{"approved": fmt.Sprint(approved), "rejected": fmt.Sprint(len(records) - approved)}}); err != nil {
		return nil, err
	}
	return records, nil
}

func (s *Service) ReopenRejected(batchID, reviewer string) (int, error) {
	if err := ValidateReviewer(reviewer); err != nil {
		return 0, err
	}
	items, err := s.store.ListItemsByStatus(batchID, domain.ItemRejected)
	if err != nil {
		return 0, err
	}
	for _, item := range items {
		if err := s.store.UpdateItemStatus(item.ID, domain.ItemPending, "reopened for correction", reviewer); err != nil {
			return 0, err
		}
	}
	if len(items) > 0 {
		if err := s.store.UpdateBatchStatus(batchID, domain.BatchSubmitted, reviewer, ""); err != nil {
			return 0, err
		}
	}
	return len(items), nil
}

func (s *Service) ReviewSingle(batchID, itemID, reviewer string) (domain.ReviewRecord, error) {
	if err := ValidateReviewer(reviewer); err != nil {
		return domain.ReviewRecord{}, err
	}
	item, err := s.store.GetItem(itemID)
	if err != nil {
		return domain.ReviewRecord{}, err
	}
	if item.BatchID != batchID {
		return domain.ReviewRecord{}, fmt.Errorf("item %s does not belong to batch %s", itemID, batchID)
	}
	decision, reason := s.policy.Evaluate(item)
	status := domain.ItemRejected
	if decision == domain.DecisionApprove {
		status = domain.ItemApproved
	}
	if err := s.store.UpdateItemStatus(item.ID, status, reason, reviewer); err != nil {
		return domain.ReviewRecord{}, err
	}
	record := domain.ReviewRecord{ID: fmt.Sprintf("review-%s-%s", batchID, itemID), BatchID: batchID, ItemID: itemID, Reviewer: reviewer, Decision: decision, Reason: reason, Sequence: item.Position}
	if err := s.store.SaveReview(record); err != nil {
		return domain.ReviewRecord{}, err
	}
	return record, nil
}
