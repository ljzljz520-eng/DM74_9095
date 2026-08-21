package workflow39

import (
	"fmt"
	"sort"
	"strings"

	"theatre39/internal/domain"
)

type AuditEntry struct {
	Sequence int
	Kind     string
	Message  string
	Actor    string
}

type AuditTrail struct {
	BatchID string
	Entries []AuditEntry
}

func (c *Coordinator) BuildAuditTrail(batchID string) (AuditTrail, error) {
	events, err := c.Store.ListEvents(batchID)
	if err != nil {
		return AuditTrail{}, err
	}
	trail := AuditTrail{BatchID: batchID, Entries: make([]AuditEntry, 0, len(events))}
	for index, event := range events {
		actor := event.Attributes["actor"]
		if actor == "" {
			actor = event.Attributes["owner"]
		}
		trail.Entries = append(trail.Entries, AuditEntry{Sequence: index + 1, Kind: event.Kind, Message: event.Message, Actor: actor})
	}
	return trail, nil
}

func (c *Coordinator) ValidateBatchClosure(batchID string) error {
	batch, err := c.Store.GetBatch(batchID)
	if err != nil {
		return err
	}
	items, err := c.Store.ListItems(batchID)
	if err != nil {
		return err
	}
	if batch.Status == domain.BatchArchived {
		for _, item := range items {
			if item.Status != domain.ItemCompleted && item.Status != domain.ItemRejected {
				return fmt.Errorf("archived batch has open item %s", item.ID)
			}
		}
		return nil
	}
	if batch.Status == domain.BatchPartial {
		for _, item := range items {
			if item.Status == domain.ItemPending {
				return nil
			}
		}
		return fmt.Errorf("partial batch %s has no pending follow-up", batchID)
	}
	return fmt.Errorf("batch %s is not closed", batchID)
}

func (t AuditTrail) Kinds() []string {
	kinds := make([]string, 0, len(t.Entries))
	seen := make(map[string]bool)
	for _, entry := range t.Entries {
		if !seen[entry.Kind] {
			seen[entry.Kind] = true
			kinds = append(kinds, entry.Kind)
		}
	}
	sort.Strings(kinds)
	return kinds
}

func (t AuditTrail) Render() string {
	lines := make([]string, 0, len(t.Entries))
	for _, entry := range t.Entries {
		actor := entry.Actor
		if actor == "" {
			actor = "system"
		}
		lines = append(lines, fmt.Sprintf("%02d %s [%s] %s", entry.Sequence, actor, entry.Kind, entry.Message))
	}
	return strings.Join(lines, "\n")
}

func (c *Coordinator) ReleaseBatch(batchID, reason string) (int, error) {
	count, err := c.Store.ReleaseBatchHolds(batchID, reason)
	if err != nil {
		return 0, err
	}
	if err := c.Store.RecordEvent(domain.BatchEvent{ID: "event-release-" + batchID, BatchID: batchID, Kind: "holds-released", Message: fmt.Sprintf("released %d holds", count)}); err != nil {
		return 0, err
	}
	return count, nil
}
