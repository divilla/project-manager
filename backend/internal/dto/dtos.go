package dto

type (
	// ChangePhase defines ChangePhase values.
	ChangePhase struct {
		Slug     string `json:"slug"`
		Priority int    `json:"priority"`
	}

	// ChangeType defines ChangeType values.
	ChangeType struct {
		Slug     string `json:"slug"`
		Priority int    `json:"priority"`
	}
)
