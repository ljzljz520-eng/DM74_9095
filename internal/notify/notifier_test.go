package notify_test

import (
	"path/filepath"
	"testing"

	"theatre39/internal/domain"
	"theatre39/internal/notify"
	"theatre39/internal/store"
)

func TestNotifierClassifiesOutcomes(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "notify.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	notifier := notify.New(db)
	message, err := notifier.Publish(domain.BatchSummary{BatchID: "batch-1", Total: 2, Approved: 1, Rejected: 1, FailedItemIDs: []string{"item-2"}}, notify.ChannelDesk, "desk")
	if err != nil {
		t.Fatal(err)
	}
	if !notify.IsPartial(message) {
		t.Fatalf("mixed outcomes must publish partial notification, got %+v", message)
	}
	if len(message.Body) == 0 {
		t.Fatal("message body is empty")
	}
}
