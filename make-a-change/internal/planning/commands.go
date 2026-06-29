package planning

// MainCommands returns slash commands for the main screen.
func MainCommands() []string {
	return []string{"/new-change", "/changes", "/epics", "/projects", "/select-project", "/help", "/quit"}
}
