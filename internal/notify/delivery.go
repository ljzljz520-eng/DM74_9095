package notify

import (
	"fmt"
	"sort"
	"strings"

	"theatre39/internal/domain"
)

type DeliveryReceipt struct {
	MessageID string
	Channel   Channel
	Recipient string
	Accepted  bool
	Attempt   int
	Error     string
}

type DeliveryPlan struct {
	BatchID  string
	Messages []Message
	Receipts []DeliveryReceipt
}

func NewPlan(summary domain.BatchSummary, recipients map[Channel]string) DeliveryPlan {
	channels := make([]Channel, 0, len(recipients))
	for channel := range recipients {
		channels = append(channels, channel)
	}
	sort.Slice(channels, func(i, j int) bool { return channels[i] < channels[j] })
	messages := make([]Message, 0, len(channels))
	for _, channel := range channels {
		messages = append(messages, Render(summary, channel, recipients[channel]))
	}
	return DeliveryPlan{BatchID: summary.BatchID, Messages: messages, Receipts: make([]DeliveryReceipt, 0, len(messages))}
}

func (p *DeliveryPlan) Accept(index int, attempt int) error {
	if index < 0 || index >= len(p.Messages) {
		return fmt.Errorf("message index %d is out of range", index)
	}
	if attempt < 1 {
		return fmt.Errorf("delivery attempt must be positive")
	}
	message := p.Messages[index]
	p.Receipts = append(p.Receipts, DeliveryReceipt{MessageID: fmt.Sprintf("%s-%d", p.BatchID, index+1), Channel: message.Channel, Recipient: strings.Join(message.Recipients, ","), Accepted: true, Attempt: attempt})
	return nil
}

func (p *DeliveryPlan) Reject(index int, attempt int, reason string) error {
	if index < 0 || index >= len(p.Messages) {
		return fmt.Errorf("message index %d is out of range", index)
	}
	if attempt < 1 {
		return fmt.Errorf("delivery attempt must be positive")
	}
	if strings.TrimSpace(reason) == "" {
		return fmt.Errorf("delivery rejection needs a reason")
	}
	message := p.Messages[index]
	p.Receipts = append(p.Receipts, DeliveryReceipt{MessageID: fmt.Sprintf("%s-%d", p.BatchID, index+1), Channel: message.Channel, Recipient: strings.Join(message.Recipients, ","), Attempt: attempt, Error: reason})
	return nil
}

func (p DeliveryPlan) Complete() bool {
	if len(p.Messages) == 0 || len(p.Receipts) < len(p.Messages) {
		return false
	}
	for _, receipt := range p.Receipts {
		if !receipt.Accepted {
			return false
		}
	}
	return true
}

func (p DeliveryPlan) Retryable() []DeliveryReceipt {
	latest := make(map[string]DeliveryReceipt)
	for _, receipt := range p.Receipts {
		latest[receipt.MessageID] = receipt
	}
	retry := make([]DeliveryReceipt, 0)
	for _, receipt := range latest {
		if !receipt.Accepted && receipt.Attempt < 3 {
			retry = append(retry, receipt)
		}
	}
	sort.Slice(retry, func(i, j int) bool { return retry[i].MessageID < retry[j].MessageID })
	return retry
}

func (p DeliveryPlan) Digest() string {
	accepted := 0
	failed := 0
	for _, receipt := range p.Receipts {
		if receipt.Accepted {
			accepted++
		} else {
			failed++
		}
	}
	return fmt.Sprintf("batch=%s messages=%d accepted=%d failed=%d complete=%t", p.BatchID, len(p.Messages), accepted, failed, p.Complete())
}
