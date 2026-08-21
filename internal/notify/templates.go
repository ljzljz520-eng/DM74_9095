package notify

import (
	"fmt"
	"strings"

	"theatre39/internal/domain"
)

type Channel string

const (
	ChannelDesk  Channel = "desk"
	ChannelEmail Channel = "email"
	ChannelLog   Channel = "log"
)

type Message struct {
	BatchID    string
	Channel    Channel
	Subject    string
	Body       string
	Kind       string
	Recipients []string
}

func Render(summary domain.BatchSummary, channel Channel, recipient string) Message {
	label := strings.ToUpper(summary.Kind)
	body := fmt.Sprintf("Batch %s: %s. approved=%d rejected=%d pending=%d", summary.BatchID, label, summary.Approved, summary.Rejected, summary.Pending)
	if len(summary.FailedItemIDs) > 0 {
		body += "; failed=" + strings.Join(summary.FailedItemIDs, ",")
	}
	return Message{BatchID: summary.BatchID, Channel: channel, Subject: "Theatre reservation " + label, Body: body, Kind: summary.Kind, Recipients: []string{recipient}}
}

func IsSuccess(message Message) bool {
	return message.Kind == "success" || message.Kind == "complete"
}

func IsPartial(message Message) bool {
	return message.Kind == "partial"
}
