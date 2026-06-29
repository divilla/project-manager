package changes

// ListCommands returns slash commands for the changes list screen.
func ListCommands() []string {
	return []string{"/new-change", "/phase-filter", "/epic-filter", "/type-filter", "/find-filter", "/clear-filters", "/help", "/return"}
}

// DetailCommands returns slash commands for change details.
func DetailCommands() []string {
	return []string{"/new-requirement", "/phase", "/epic", "/types", "/edit", "/delete", "/return"}
}
