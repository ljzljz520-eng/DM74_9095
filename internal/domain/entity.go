package domain

const (
	BatchDraft      = "draft"
	BatchSubmitted  = "submitted"
	BatchReviewed   = "reviewed"
	BatchPartial    = "partial"
	BatchArchived   = "archived"
	ItemPending     = "pending"
	ItemApproved    = "approved"
	ItemRejected    = "rejected"
	ItemCompleted   = "completed"
	HoldActive      = "active"
	HoldReleased    = "released"
	DecisionApprove = "approve"
	DecisionReject  = "reject"
)

type ReservationBatch struct {
	ID            string   `json:"id"`
	Theatre       string   `json:"theatre"`
	Performance   string   `json:"performance"`
	CreatedBy     string   `json:"created_by"`
	Status        string   `json:"status"`
	ItemIDs       []string `json:"item_ids"`
	SubmittedNote string   `json:"submitted_note"`
	ReviewedBy    string   `json:"reviewed_by"`
	ArchiveID     string   `json:"archive_id"`
	Version       int      `json:"version"`
}

type ReservationItem struct {
	ID             string `json:"id"`
	BatchID        string `json:"batch_id"`
	Position       int    `json:"position"`
	Patron         string `json:"patron"`
	SeatCode       string `json:"seat_code"`
	RequestedClass string `json:"requested_class"`
	Status         string `json:"status"`
	Reason         string `json:"reason"`
	ReviewedBy     string `json:"reviewed_by"`
}

type SeatHold struct {
	ID        string `json:"id"`
	BatchID   string `json:"batch_id"`
	ItemID    string `json:"item_id"`
	SeatCode  string `json:"seat_code"`
	State     string `json:"state"`
	ExpiresOn string `json:"expires_on"`
}

type ReviewRecord struct {
	ID       string `json:"id"`
	BatchID  string `json:"batch_id"`
	ItemID   string `json:"item_id"`
	Reviewer string `json:"reviewer"`
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
	Sequence int    `json:"sequence"`
}

type ArchiveEntry struct {
	ID          string   `json:"id"`
	BatchID     string   `json:"batch_id"`
	ArchivedBy  string   `json:"archived_by"`
	ArchivedOn  string   `json:"archived_on"`
	ItemIDs     []string `json:"item_ids"`
	RejectedIDs []string `json:"rejected_ids"`
	Summary     string   `json:"summary"`
}

type BatchEvent struct {
	ID         string            `json:"id"`
	BatchID    string            `json:"batch_id"`
	Kind       string            `json:"kind"`
	Message    string            `json:"message"`
	Attributes map[string]string `json:"attributes"`
}

type ItemInput struct {
	ID             string
	Patron         string
	SeatCode       string
	RequestedClass string
}

type BatchInput struct {
	ID            string
	Theatre       string
	Performance   string
	CreatedBy     string
	SubmittedNote string
}

type ReviewOutcome struct {
	ItemID   string
	Decision string
	Reason   string
}

type BatchSummary struct {
	BatchID       string
	Total         int
	Approved      int
	Rejected      int
	Pending       int
	Kind          string
	FailedItemIDs []string
}

type Page[T any] struct {
	Items      []T
	NextCursor string
	Total      int
}
