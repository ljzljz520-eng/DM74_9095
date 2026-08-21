package review

import (
	"fmt"
	"sort"
	"strings"

	"theatre39/internal/domain"
)

type Escalation struct {
	ItemID   string
	BatchID  string
	Priority string
	Reason   string
	Owner    string
}

func BuildEscalations(items []domain.ReservationItem, owner string) []Escalation {
	escalations := make([]Escalation, 0)
	for _, item := range items {
		if item.Status != domain.ItemRejected {
			continue
		}
		priority := "normal"
		if item.Position == 5 {
			priority = "high"
		}
		if strings.Contains(strings.ToLower(item.Reason), "access") {
			priority = "urgent"
		}
		escalations = append(escalations, Escalation{ItemID: item.ID, BatchID: item.BatchID, Priority: priority, Reason: item.Reason, Owner: owner})
	}
	sort.SliceStable(escalations, func(i, j int) bool {
		if escalations[i].Priority == escalations[j].Priority {
			return escalations[i].ItemID < escalations[j].ItemID
		}
		return escalationRank(escalations[i].Priority) > escalationRank(escalations[j].Priority)
	})
	return escalations
}

func escalationRank(priority string) int {
	switch priority {
	case "urgent":
		return 3
	case "high":
		return 2
	default:
		return 1
	}
}

func (s *Service) EscalateRejected(batchID, owner string) ([]Escalation, error) {
	items, err := s.store.ListItemsByStatus(batchID, domain.ItemRejected)
	if err != nil {
		return nil, err
	}
	escalations := BuildEscalations(items, owner)
	for index, escalation := range escalations {
		if err := s.store.RecordEvent(domain.BatchEvent{ID: fmt.Sprintf("event-escalate-%s-%02d", batchID, index+1), BatchID: batchID, Kind: "escalated", Message: escalation.ItemID + ": " + escalation.Reason, Attributes: map[string]string{"priority": escalation.Priority, "owner": owner}}); err != nil {
			return nil, err
		}
	}
	return escalations, nil
}
