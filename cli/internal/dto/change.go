package dto

// Change is the change row and detail data used by mch.
type Change struct {
	ID          string
	Ref         string
	Slug        string
	ProjectID   string
	EpicID      string
	EpicName    string
	ChangePhase string
	ChangeTypes []string
	Title       string
	Body        string
	PRBody      string
	PRUrl       string
	AgentEdit   bool
	Open        bool
	Done        int
	Total       int
	Completed   int
	Created     string
	Modified    string
}

// ChangeCreateInput is the backend payload for creating a change.
type ChangeCreateInput struct {
	ProjectID   int
	Title       string
	Body        string
	ChangeTypes []string
	EpicID      *int
}
