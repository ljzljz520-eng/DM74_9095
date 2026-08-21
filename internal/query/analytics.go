package query

import (
	"fmt"
	"sort"
	"strings"

	"theatre39/internal/domain"
)

type TheatreMetric struct {
	Theatre     string
	Batches     int
	Items       int
	Approved    int
	Rejected    int
	Pending     int
	ActiveHolds int
}

func (s *Service) Metrics(filter BatchFilter) ([]TheatreMetric, error) {
	page, err := s.FindBatches(filter)
	if err != nil {
		return nil, err
	}
	byTheatre := make(map[string]*TheatreMetric)
	for _, batch := range page.Items {
		metric := byTheatre[batch.Theatre]
		if metric == nil {
			metric = &TheatreMetric{Theatre: batch.Theatre}
			byTheatre[batch.Theatre] = metric
		}
		metric.Batches++
		items, listErr := s.store.ListItems(batch.ID)
		if listErr != nil {
			return nil, listErr
		}
		metric.Items += len(items)
		for _, item := range items {
			switch item.Status {
			case domain.ItemApproved, domain.ItemCompleted:
				metric.Approved++
			case domain.ItemRejected:
				metric.Rejected++
			case domain.ItemPending:
				metric.Pending++
			}
		}
		holds, holdErr := s.store.ListHolds(batch.ID)
		if holdErr != nil {
			return nil, holdErr
		}
		for _, hold := range holds {
			if hold.State == domain.HoldActive {
				metric.ActiveHolds++
			}
		}
	}
	metrics := make([]TheatreMetric, 0, len(byTheatre))
	for _, metric := range byTheatre {
		metrics = append(metrics, *metric)
	}
	sort.Slice(metrics, func(i, j int) bool { return metrics[i].Theatre < metrics[j].Theatre })
	return metrics, nil
}

func (m TheatreMetric) CompletionRate() float64 {
	if m.Items == 0 {
		return 0
	}
	return float64(m.Approved) / float64(m.Items)
}

func (m TheatreMetric) Label() string {
	return fmt.Sprintf("%s batches=%d items=%d approved=%d rejected=%d pending=%d active_holds=%d rate=%.2f", m.Theatre, m.Batches, m.Items, m.Approved, m.Rejected, m.Pending, m.ActiveHolds, m.CompletionRate())
}

func (s *Service) SeatOccupancy(batchID string) (map[string]bool, error) {
	holds, err := s.store.ListHolds(batchID)
	if err != nil {
		return nil, err
	}
	occupied := make(map[string]bool, len(holds))
	for _, hold := range holds {
		occupied[strings.ToUpper(hold.SeatCode)] = hold.State == domain.HoldActive
	}
	return occupied, nil
}

func (s *Service) StatusCounts(batchID string) (map[string]int, error) {
	items, err := s.store.ListItems(batchID)
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int)
	for _, item := range items {
		counts[item.Status]++
	}
	return counts, nil
}
