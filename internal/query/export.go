package query

import (
	"encoding/csv"
	"fmt"
	"sort"
	"strings"

	"theatre39/internal/domain"
)

type ItemResult struct {
	BatchID    string
	BatchState string
	Item       domain.ReservationItem
}

func (s *Service) SearchItems(filter BatchFilter) ([]ItemResult, error) {
	filter = filter.Normalize()
	page, err := s.FindBatches(filter)
	if err != nil {
		return nil, err
	}
	results := make([]ItemResult, 0)
	for _, batch := range page.Items {
		items, listErr := s.store.ListItems(batch.ID)
		if listErr != nil {
			return nil, listErr
		}
		for _, item := range items {
			if matchesItem(item, filter) {
				results = append(results, ItemResult{BatchID: batch.ID, BatchState: batch.Status, Item: item})
			}
		}
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].BatchID == results[j].BatchID {
			return results[i].Item.Position < results[j].Item.Position
		}
		return results[i].BatchID < results[j].BatchID
	})
	return results, nil
}

func (s *Service) ExportItemsCSV(filter BatchFilter) (string, error) {
	items, err := s.SearchItems(filter)
	if err != nil {
		return "", err
	}
	var builder strings.Builder
	writer := csv.NewWriter(&builder)
	if err := writer.Write([]string{"batch_id", "batch_state", "item_id", "position", "patron", "seat", "class", "status", "reason"}); err != nil {
		return "", err
	}
	for _, result := range items {
		row := []string{result.BatchID, result.BatchState, result.Item.ID, fmt.Sprint(result.Item.Position), result.Item.Patron, result.Item.SeatCode, result.Item.RequestedClass, result.Item.Status, result.Item.Reason}
		if err := writer.Write(row); err != nil {
			return "", err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}
	return builder.String(), nil
}

func (s *Service) ReviewTimeline(batchID string) ([]domain.ReviewRecord, error) {
	reviews, err := s.store.ListReviews(batchID)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(reviews, func(i, j int) bool { return reviews[i].Sequence < reviews[j].Sequence })
	return reviews, nil
}

func FormatItemResult(result ItemResult) string {
	return fmt.Sprintf("%s %s %s seat=%s status=%s", result.BatchID, result.Item.ID, result.Item.Patron, result.Item.SeatCode, result.Item.Status)
}
