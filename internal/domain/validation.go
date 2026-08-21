package domain

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var seatPattern = regexp.MustCompile(`^[A-Z]{1,2}-[0-9]{1,3}$`)

func ValidateBatchInput(input BatchInput) error {
	if strings.TrimSpace(input.ID) == "" {
		return fmt.Errorf("batch id is required")
	}
	if strings.TrimSpace(input.Theatre) == "" {
		return fmt.Errorf("theatre is required")
	}
	if strings.TrimSpace(input.Performance) == "" {
		return fmt.Errorf("performance is required")
	}
	if strings.TrimSpace(input.CreatedBy) == "" {
		return fmt.Errorf("creator is required")
	}
	return nil
}

func ValidateItemInput(input ItemInput) error {
	if strings.TrimSpace(input.ID) == "" {
		return fmt.Errorf("item id is required")
	}
	if strings.TrimSpace(input.Patron) == "" {
		return fmt.Errorf("patron is required")
	}
	if !seatPattern.MatchString(strings.ToUpper(strings.TrimSpace(input.SeatCode))) {
		return fmt.Errorf("seat code %q is invalid", input.SeatCode)
	}
	if strings.TrimSpace(input.RequestedClass) == "" {
		return fmt.Errorf("requested class is required")
	}
	return nil
}

func ValidateItemSet(items []ItemInput) error {
	if len(items) == 0 {
		return fmt.Errorf("at least one item is required")
	}
	seenIDs := make(map[string]struct{}, len(items))
	seenSeats := make(map[string]struct{}, len(items))
	for index, item := range items {
		if err := ValidateItemInput(item); err != nil {
			return fmt.Errorf("item %d: %w", index+1, err)
		}
		id := strings.TrimSpace(item.ID)
		seat := strings.ToUpper(strings.TrimSpace(item.SeatCode))
		if _, exists := seenIDs[id]; exists {
			return fmt.Errorf("duplicate item id %s", id)
		}
		if _, exists := seenSeats[seat]; exists {
			return fmt.Errorf("duplicate seat %s", seat)
		}
		seenIDs[id] = struct{}{}
		seenSeats[seat] = struct{}{}
	}
	return nil
}

func NormalizeItemInput(input ItemInput) ItemInput {
	input.ID = strings.TrimSpace(input.ID)
	input.Patron = strings.TrimSpace(input.Patron)
	input.SeatCode = strings.ToUpper(strings.TrimSpace(input.SeatCode))
	input.RequestedClass = strings.ToLower(strings.TrimSpace(input.RequestedClass))
	return input
}

func SortItems(items []ReservationItem) []ReservationItem {
	copyItems := append([]ReservationItem(nil), items...)
	sort.SliceStable(copyItems, func(i, j int) bool {
		if copyItems[i].Position == copyItems[j].Position {
			return copyItems[i].ID < copyItems[j].ID
		}
		return copyItems[i].Position < copyItems[j].Position
	})
	return copyItems
}

func IsTerminalItem(status string) bool {
	return status == ItemCompleted || status == ItemRejected
}

func IsOpenBatch(status string) bool {
	return status == BatchDraft || status == BatchSubmitted || status == BatchReviewed || status == BatchPartial
}
