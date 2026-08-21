package notify

import (
	"fmt"

	"theatre39/internal/domain"
	"theatre39/internal/store"
)

type Notifier struct {
	store *store.DB
}

func New(db *store.DB) *Notifier {
	return &Notifier{store: db}
}

func (n *Notifier) Compose(summary domain.BatchSummary) domain.BatchSummary {
	if summary.Total == 0 {
		summary.Kind = "empty"
		return summary
	}
	if summary.Rejected == 0 && summary.Pending == 0 {
		summary.Kind = "success"
		return summary
	}
	if summary.Rejected > 0 && summary.Approved > 0 {
		summary.Kind = "partial"
		return summary
	}
	if summary.Rejected == summary.Total {
		summary.Kind = "failure"
		return summary
	}
	summary.Kind = "pending"
	return summary
}

func (n *Notifier) Publish(summary domain.BatchSummary, channel Channel, recipient string) (Message, error) {
	result := n.Compose(summary)
	message := Render(result, channel, recipient)
	event := domain.BatchEvent{ID: fmt.Sprintf("event-notify-%s-%s", summary.BatchID, channel), BatchID: summary.BatchID, Kind: "notification", Message: message.Body, Attributes: map[string]string{"channel": string(channel), "result": message.Kind}}
	if err := n.store.RecordEvent(event); err != nil {
		return Message{}, err
	}
	return message, nil
}

func (n *Notifier) PublishAll(summary domain.BatchSummary, recipients []string) ([]Message, error) {
	messages := make([]Message, 0, len(recipients))
	for index, recipient := range recipients {
		channel := ChannelLog
		if index == 0 {
			channel = ChannelDesk
		}
		if index == 1 {
			channel = ChannelEmail
		}
		message, err := n.Publish(summary, channel, recipient)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	return messages, nil
}
