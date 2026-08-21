package workflow39

import (
	"fmt"

	"theatre39/internal/archive"
	"theatre39/internal/domain"
	"theatre39/internal/intake"
	"theatre39/internal/notify"
	"theatre39/internal/query"
	"theatre39/internal/review"
	"theatre39/internal/store"
)

type Coordinator struct {
	Store   *store.DB
	Intake  *intake.Service
	Review  *review.Service
	Query   *query.Service
	Archive *archive.Service
	Notify  *notify.Notifier
}

func New(db *store.DB) *Coordinator {
	return &Coordinator{
		Store:   db,
		Intake:  intake.NewService(db),
		Review:  review.NewService(db, review.DefaultPolicy()),
		Query:   query.NewService(db),
		Archive: archive.NewService(db, archive.DefaultRetentionPolicy()),
		Notify:  notify.New(db),
	}
}

func (c *Coordinator) RegisterAndSubmit(input domain.BatchInput, items []domain.ItemInput) (domain.ReservationBatch, error) {
	batch, err := c.Intake.Register(input, items)
	if err != nil {
		return domain.ReservationBatch{}, err
	}
	if err := c.Intake.Submit(batch.ID, input.SubmittedNote); err != nil {
		return domain.ReservationBatch{}, err
	}
	return c.Store.GetBatch(batch.ID)
}

func (c *Coordinator) ReviewAndNotify(batchID, reviewer, recipient string) (Report, error) {
	report := NewReport(batchID)
	report.Stage = "review"
	if _, err := c.Review.ReviewBatch(batchID, reviewer); err != nil {
		return report, err
	}
	summary, err := c.Query.Summarize(batchID)
	if err != nil {
		return report, err
	}
	report.ApplySummary(summary)
	report.Stage = "notification"
	message, err := c.Notify.Publish(summary, notify.ChannelDesk, recipient)
	if err != nil {
		return report, err
	}
	report.Notification = message.Kind
	if notify.IsSuccess(message) {
		if err := c.markItemsCompleted(batchID, reviewer); err != nil {
			return report, err
		}
	}
	items, err := c.Store.ListItems(batchID)
	if err != nil {
		return report, err
	}
	report.Completed = CountCompleted(items)
	return report, nil
}

func (c *Coordinator) markItemsCompleted(batchID, actor string) error {
	items, err := c.Store.ListItems(batchID)
	if err != nil {
		return err
	}
	for _, item := range items {
		if item.Status == domain.ItemCompleted {
			continue
		}
		if err := c.Store.UpdateItemStatus(item.ID, domain.ItemCompleted, "notification accepted", actor); err != nil {
			return err
		}
	}
	return c.Store.UpdateBatchStatus(batchID, domain.BatchReviewed, actor, "")
}

func (c *Coordinator) ArchiveReviewed(batchID, actor, archivedOn, note string) (domain.ArchiveEntry, error) {
	entry, err := c.Archive.Archive(batchID, actor, archivedOn, note)
	if err != nil {
		return domain.ArchiveEntry{}, err
	}
	return entry, nil
}

func (c *Coordinator) FullChain(input domain.BatchInput, items []domain.ItemInput, reviewer, recipient, archiveOn, note string) (Report, error) {
	batch, err := c.RegisterAndSubmit(input, items)
	if err != nil {
		return Report{}, err
	}
	report, err := c.ReviewAndNotify(batch.ID, reviewer, recipient)
	if err != nil {
		return report, err
	}
	if report.Rejected == 0 {
		entry, archiveErr := c.ArchiveReviewed(batch.ID, reviewer, archiveOn, note)
		if archiveErr != nil {
			return report, archiveErr
		}
		report.ArchiveID = entry.ID
		report.Stage = "archive"
	} else {
		report.Warnings = append(report.Warnings, fmt.Sprintf("%d items require follow-up", report.Rejected))
	}
	return report, nil
}

func (c *Coordinator) ReopenAndReview(batchID, reviewer string) (Report, error) {
	report := NewReport(batchID)
	count, err := c.Review.ReopenRejected(batchID, reviewer)
	if err != nil {
		return report, err
	}
	report.Stage = "reopen"
	report.Status = fmt.Sprintf("%d reopened", count)
	return report, nil
}
