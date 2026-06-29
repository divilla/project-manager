package requirements

// DetailCommands returns slash commands for requirement details.
func DetailCommands() []string {
	return []string{"/new-requirement", "/edit", "/delete", "/save", "/cancel", "/return"}
}

// EditCommands returns slash commands for requirement edit/create screens.
func EditCommands() []string {
	return []string{"/save", "/cancel"}
}
