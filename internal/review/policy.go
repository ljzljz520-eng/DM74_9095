package review

import (
	"fmt"
	"strings"

	"theatre39/internal/domain"
)

type Policy struct {
	AllowedClasses map[string]bool
	BlockedSeats   map[string]bool
	RequiredPrefix string
}

func DefaultPolicy() Policy {
	return Policy{
		AllowedClasses: map[string]bool{"standard": true, "accessible": true, "student": true, "senior": true},
		BlockedSeats:   map[string]bool{"A-01": true, "A-02": true},
		RequiredPrefix: "",
	}
}

func (p Policy) Evaluate(item domain.ReservationItem) (string, string) {
	seat := strings.ToUpper(strings.TrimSpace(item.SeatCode))
	className := strings.ToLower(strings.TrimSpace(item.RequestedClass))
	if strings.TrimSpace(item.Patron) == "" {
		return domain.DecisionReject, "patron is missing"
	}
	if p.RequiredPrefix != "" && !strings.HasPrefix(item.Patron, p.RequiredPrefix) {
		return domain.DecisionReject, "patron is outside community"
	}
	if !p.AllowedClasses[className] {
		return domain.DecisionReject, "requested class is not supported"
	}
	if p.BlockedSeats[seat] {
		return domain.DecisionReject, "seat is reserved for operational access"
	}
	if item.Position <= 0 {
		return domain.DecisionReject, "item position is invalid"
	}
	return domain.DecisionApprove, "policy checks passed"
}

func ValidateReviewer(reviewer string) error {
	if strings.TrimSpace(reviewer) == "" {
		return fmt.Errorf("reviewer is required")
	}
	if len([]rune(reviewer)) < 2 {
		return fmt.Errorf("reviewer name is too short")
	}
	return nil
}

func ExplainDecision(decision, reason string) string {
	if decision == domain.DecisionApprove {
		return "approved: " + reason
	}
	return "rejected: " + reason
}
