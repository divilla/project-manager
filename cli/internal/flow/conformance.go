package flow

// ConformanceDefinition returns a fresh definition covering every supported mode,
// artifact-local Editor Steps, and same-Step Chat editing.
func ConformanceDefinition() Definition {
	const (
		ideaEdit        StepID   = "idea-edit"
		specEdit        StepID   = "spec-edit"
		specReview      StepID   = "spec-review"
		prEdit          StepID   = "pr-edit"
		prChat          StepID   = "pr-chat"
		errorScreen     ScreenID = "error"
		ideaEditor      ScreenID = "idea-editor"
		specEditor      ScreenID = "spec-editor"
		specExec        ScreenID = "spec-exec"
		specChat        ScreenID = "spec-chat"
		prEditor        ScreenID = "pr-editor"
		prChatScreen    ScreenID = "pr-chat"
		ideaPreview     ScreenID = "idea-preview"
		specEditPreview ScreenID = "spec-edit-preview"
		specPreview     ScreenID = "spec-preview"
		prEditPreview   ScreenID = "pr-edit-preview"
		prPreview       ScreenID = "pr-preview"
	)
	return Definition{
		ID: "conformance",
		Steps: []StepDefinition{
			{ID: ideaEdit, Mode: ModeEditor, Tasks: []TaskDefinition{{
				ID: "edit-idea", Type: TaskEditor, Artifact: ArtifactIdea, Screen: ideaEditor,
				Preview: ideaPreview, Cancel: ScreenDestination(ChangeDetailsTerminal), Error: errorScreen,
			}}},
			{ID: specEdit, Mode: ModeEditor, Tasks: []TaskDefinition{{
				ID: "edit-spec", Type: TaskEditor, Artifact: ArtifactSpec, Screen: specEditor,
				Preview: specEditPreview, Cancel: ScreenDestination(ChangeDetailsTerminal), Error: errorScreen,
			}}},
			{ID: specReview, Mode: ModeExec, Tasks: []TaskDefinition{
				{
					ID: "run-spec-review", Type: TaskExec, Artifact: ArtifactSpec, Screen: specExec,
					Prompt: "prompts/spec-review.md", Script: "scripts/codex-exec-restore-session.sh",
					ExpectedOutput: "No blocking issues found.", UnexpectedOutput: specChat,
					Preview: specPreview, Error: errorScreen,
				},
				{
					ID: "clarify-spec", Type: TaskChat, Artifact: ArtifactSpec, Screen: specChat,
					Script: "scripts/codex-resume-session.sh", Editor: ScreenDestination(specEditor),
					Preview: specPreview, Error: errorScreen,
				},
			}},
			{ID: prEdit, Mode: ModeEditor, Tasks: []TaskDefinition{{
				ID: "edit-pr", Type: TaskEditor, Artifact: ArtifactPR, Screen: prEditor,
				Preview: prEditPreview, Cancel: ScreenDestination(ChangeDetailsTerminal), Error: errorScreen,
			}}},
			{ID: prChat, Mode: ModeChat, Tasks: []TaskDefinition{{
				ID: "clarify-pr", Type: TaskChat, Artifact: ArtifactPR, Screen: prChatScreen,
				Script: "scripts/codex-resume-session.sh", Editor: ScreenDestination(prEditor),
				Preview: prPreview, Error: errorScreen,
			}}},
		},
		Screens: []ScreenDefinition{
			{ID: errorScreen, Type: ScreenError, Title: "Flow Error"},
			{ID: ideaEditor, Type: ScreenEditor, Title: "Idea Editor"},
			{ID: specEditor, Type: ScreenEditor, Title: "Spec Editor"},
			{ID: specExec, Type: ScreenExec, Title: "Spec Review"},
			{ID: specChat, Type: ScreenChat, Title: "Spec Chat"},
			{ID: prEditor, Type: ScreenEditor, Title: "PR Editor"},
			{ID: prChatScreen, Type: ScreenChat, Title: "PR Chat"},
			conformancePreview(ideaPreview, ideaEdit, StepDestination(specEdit), StepDestination(ideaEdit)),
			conformancePreview(specEditPreview, specEdit, StepDestination(specReview), StepDestination(specEdit)),
			conformancePreview(specPreview, specReview, StepDestination(prEdit), StepDestination(specEdit)),
			conformancePreview(prEditPreview, prEdit, StepDestination(prChat), StepDestination(prEdit)),
			conformancePreview(prPreview, prChat, ScreenDestination(MainTerminal), StepDestination(prEdit)),
		},
	}
}

func conformancePreview(id ScreenID, from StepID, next Destination, edit Destination) ScreenDefinition {
	return ScreenDefinition{
		ID: id, Type: ScreenPreview, FromStep: from,
		Commands: []CommandDefinition{
			{ID: "/continue", Destination: next},
			{ID: "/edit", Destination: edit},
			{ID: "/cancel", Destination: ScreenDestination(ChangeDetailsTerminal)},
		},
	}
}
