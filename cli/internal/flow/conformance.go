package flow

// ConformanceDefinition returns a fresh definition that exercises every initial task shape.
func ConformanceDefinition() Definition {
	const (
		errorScreen       ScreenID = "error"
		editorScreen      ScreenID = "editor"
		execScreen        ScreenID = "exec"
		interactiveScreen ScreenID = "interactive"
		ideaPreview       ScreenID = "idea-preview"
		specPreview       ScreenID = "spec-preview"
		prPreview         ScreenID = "pr-preview"
		changesScreen     ScreenID = "changes"
	)
	return Definition{
		ID: "conformance",
		Steps: []StepDefinition{
			{
				ID: "idea-edit",
				Tasks: []TaskDefinition{{
					ID:       "edit-idea",
					Type:     TaskEditor,
					Artifact: ArtifactIdea,
					Screen:   editorScreen,
					Preview:  ideaPreview,
					Error:    errorScreen,
				}},
			},
			{
				ID: "spec-review",
				Tasks: []TaskDefinition{
					{
						ID:               "run-spec-review",
						Type:             TaskExec,
						Artifact:         ArtifactSpec,
						Screen:           execScreen,
						Prompt:           "prompts/spec-review.md",
						Script:           "scripts/codex-exec-restore-session.sh",
						ExpectedOutput:   "No blocking issues found.",
						Preview:          specPreview,
						UnexpectedOutput: interactiveScreen,
						Error:            errorScreen,
					},
					{
						ID:       "clarify-spec",
						Type:     TaskInteractive,
						Artifact: ArtifactSpec,
						Screen:   interactiveScreen,
						Script:   "scripts/codex-resume-session.sh",
						Editor:   editorScreen,
						Preview:  specPreview,
						Error:    errorScreen,
					},
				},
			},
			{
				ID: "pr-review",
				Tasks: []TaskDefinition{{
					ID:       "clarify-pr",
					Type:     TaskInteractive,
					Artifact: ArtifactPR,
					Screen:   interactiveScreen,
					Script:   "scripts/codex-resume-session.sh",
					Editor:   editorScreen,
					Preview:  prPreview,
					Error:    errorScreen,
				}},
			},
		},
		Screens: []ScreenDefinition{
			{ID: errorScreen, Type: ScreenError, Title: "Flow Error"},
			{ID: editorScreen, Type: ScreenEditor, Title: "Editor"},
			{ID: execScreen, Type: ScreenExec, Title: "Running"},
			{ID: interactiveScreen, Type: ScreenInteractive, Title: "Interactive"},
			{
				ID:    ideaPreview,
				Type:  ScreenPreview,
				Title: "Idea Preview",
				Commands: []CommandDefinition{
					{ID: "/spec", Destination: StepDestination("spec-review")},
					{ID: "/return", Destination: ScreenDestination(changesScreen)},
				},
				Options: ScreenOptions{Theme: "Coldark-Dark"},
			},
			{
				ID:    specPreview,
				Type:  ScreenPreview,
				Title: "Spec Preview",
				Commands: []CommandDefinition{
					{ID: "/pr", Destination: StepDestination("pr-review")},
					{ID: "/return", Destination: ScreenDestination(changesScreen)},
				},
			},
			{
				ID:       prPreview,
				Type:     ScreenPreview,
				Title:    "PR Preview",
				Commands: []CommandDefinition{{ID: "/return", Destination: ScreenDestination(changesScreen)}},
			},
		},
	}
}
