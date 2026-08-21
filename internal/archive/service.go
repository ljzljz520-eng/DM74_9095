package archive

import (
	"fmt"

	"theatre39/internal/domain"
	"theatre39/internal/store"
)

type Service struct {
	store  *store.DB
	policy RetentionPolicy
}

func NewService(db *store.DB, policy RetentionPolicy) *Service {
	return &Service{store: db, policy: policy}
}

func (s *Service) Archive(batchID, archivedBy, archivedOn, note string) (domain.ArchiveEntry, error) {
	batch, err := s.store.GetBatch(batchID)
	if err != nil {
		return domain.ArchiveEntry{}, err
	}
	if err := s.policy.Validate(batch, note); err != nil {
		return domain.ArchiveEntry{}, err
	}
	items, err := s.store.ListItems(batchID)
	if err != nil {
		return domain.ArchiveEntry{}, err
	}
	entry := domain.ArchiveEntry{ID: "archive-" + batchID, BatchID: batchID, ArchivedBy: archivedBy, ArchivedOn: archivedOn, Summary: note}
	for _, item := range items {
		entry.ItemIDs = append(entry.ItemIDs, item.ID)
		if item.Status == domain.ItemRejected {
			entry.RejectedIDs = append(entry.RejectedIDs, item.ID)
		}
	}
	if len(entry.ItemIDs) == 0 {
		return domain.ArchiveEntry{}, fmt.Errorf("cannot archive empty batch")
	}
	if err := s.store.SaveArchive(entry); err != nil {
		return domain.ArchiveEntry{}, err
	}
	if err := s.store.UpdateBatchStatus(batchID, domain.BatchArchived, batch.ReviewedBy, entry.ID); err != nil {
		return domain.ArchiveEntry{}, err
	}
	for _, item := range items {
		if item.Status == domain.ItemApproved {
			if err := s.store.UpdateItemStatus(item.ID, domain.ItemCompleted, "archived", archivedBy); err != nil {
				return domain.ArchiveEntry{}, err
			}
		}
		if item.Status == domain.ItemRejected {
			if err := s.store.UpdateItemStatus(item.ID, domain.ItemRejected, item.Reason, archivedBy); err != nil {
				return domain.ArchiveEntry{}, err
			}
		}
	}
	if err := s.store.RecordEvent(domain.BatchEvent{ID: "event-archive-" + batchID, BatchID: batchID, Kind: "archived", Message: BuildSummary(entry)}); err != nil {
		return domain.ArchiveEntry{}, err
	}
	return entry, nil
}

func (s *Service) Restore(batchID, actor string) error {
	batch, err := s.store.GetBatch(batchID)
	if err != nil {
		return err
	}
	if batch.Status != domain.BatchArchived {
		return fmt.Errorf("batch %s is not archived", batchID)
	}
	items, err := s.store.ListItems(batchID)
	if err != nil {
		return err
	}
	for _, item := range items {
		if item.Status == domain.ItemCompleted {
			if err := s.store.UpdateItemStatus(item.ID, domain.ItemApproved, "restored", actor); err != nil {
				return err
			}
		}
	}
	if err := s.store.UpdateBatchStatus(batchID, domain.BatchReviewed, actor, ""); err != nil {
		return err
	}
	return s.store.RecordEvent(domain.BatchEvent{ID: "event-restore-" + batchID, BatchID: batchID, Kind: "restored", Message: "archive restored for correction"})
}
