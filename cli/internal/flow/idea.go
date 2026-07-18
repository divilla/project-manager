package flow

// Built-in Idea Stage definition identifiers.
const (
	IdeaDefinitionID DefinitionID = "idea-stage"

	IdeaEdit        StepID = "IdeaEdit"
	IdeaRewriteExec StepID = "IdeaRewriteExec"
	IdeaReviewExec  StepID = "IdeaReviewExec"

	IdeaEditScreen        ScreenID = "IdeaEditScreen"
	IdeaRewriteScreen     ScreenID = "IdeaRewriteExecScreen"
	IdeaRewriteChat       ScreenID = "IdeaRewriteChat"
	IdeaReviewScreen      ScreenID = "IdeaReviewExecScreen"
	IdeaReviewChat        ScreenID = "IdeaReviewChat"
	IdeaEditPreview       ScreenID = "IdeaEditPreview"
	IdeaRewritePreview    ScreenID = "IdeaRewritePreview"
	IdeaReviewPreview     ScreenID = "IdeaReviewPreview"
	GenericErrorScreen    ScreenID = "FlowErrorScreen"
	MainTerminal          ScreenID = "MainState"
	ChangesListTerminal   ScreenID = "ChangesListState"
	ChangeDetailsTerminal ScreenID = "ChangeDetailsState"
)

// IdeaDefinition returns a fresh independently mutable built-in Idea Stage definition.
func IdeaDefinition() Definition {
	return Definition{
		ID: IdeaDefinitionID,
		Steps: []StepDefinition{
			{
				ID:   IdeaEdit,
				Mode: ModeEditor,
				Tasks: []TaskDefinition{{
					ID:       "IdeaEditEditorTask",
					Type:     TaskEditor,
					Artifact: ArtifactIdea,
					Screen:   IdeaEditScreen,
					Preview:  IdeaEditPreview,
					Cancel:   ScreenDestination(ChangeDetailsTerminal),
					Error:    GenericErrorScreen,
				}},
			},
			{
				ID:   IdeaRewriteExec,
				Mode: ModeExec,
				Tasks: []TaskDefinition{
					{
						ID:               "IdeaRewriteExecTask",
						Type:             TaskExec,
						Artifact:         ArtifactIdea,
						Screen:           IdeaRewriteScreen,
						Prompt:           "prompts/idea-rewrite.md",
						Script:           "scripts/codex-exec-restore-session.sh",
						ExpectedOutput:   "Done.",
						Preview:          IdeaRewritePreview,
						UnexpectedOutput: IdeaRewriteChat,
						Error:            GenericErrorScreen,
					},
					{
						ID:       "IdeaRewriteChatTask",
						Type:     TaskChat,
						Artifact: ArtifactIdea,
						Screen:   IdeaRewriteChat,
						Script:   "scripts/codex-resume-session.sh",
						Editor:   StepDestination(IdeaEdit),
						Preview:  IdeaRewritePreview,
						Error:    GenericErrorScreen,
					},
				},
			},
			{
				ID:   IdeaReviewExec,
				Mode: ModeExec,
				Tasks: []TaskDefinition{
					{
						ID:               "IdeaReviewExecTask",
						Type:             TaskExec,
						Artifact:         ArtifactIdea,
						Screen:           IdeaReviewScreen,
						Prompt:           "prompts/idea-review.md",
						Script:           "scripts/codex-exec-restore-session.sh",
						ExpectedOutput:   "No questions or suggestions.",
						Preview:          IdeaReviewPreview,
						UnexpectedOutput: IdeaReviewChat,
						Error:            GenericErrorScreen,
					},
					{
						ID:       "IdeaReviewChatTask",
						Type:     TaskChat,
						Artifact: ArtifactIdea,
						Screen:   IdeaReviewChat,
						Script:   "scripts/codex-resume-session.sh",
						Editor:   StepDestination(IdeaEdit),
						Preview:  IdeaReviewPreview,
						Error:    GenericErrorScreen,
					},
				},
			},
		},
		Screens: []ScreenDefinition{
			{ID: GenericErrorScreen, Type: ScreenError, Title: "Flow Error"},
			{ID: IdeaEditScreen, Type: ScreenEditor, Title: "Edit Idea"},
			{ID: IdeaRewriteScreen, Type: ScreenExec, Title: "Rewrite Idea"},
			{ID: IdeaRewriteChat, Type: ScreenChat, Title: "Rewrite Chat"},
			{ID: IdeaReviewScreen, Type: ScreenExec, Title: "Review Idea"},
			{ID: IdeaReviewChat, Type: ScreenChat, Title: "Review Chat"},
			{
				ID: IdeaEditPreview, Type: ScreenPreview, Title: "Idea Preview", FromStep: IdeaEdit,
				Commands: []CommandDefinition{
					{ID: "/continue", Destination: StepDestination(IdeaRewriteExec)},
					{ID: "/edit", Destination: StepDestination(IdeaEdit)},
					{ID: "/cancel", Destination: ScreenDestination(ChangeDetailsTerminal)},
				},
			},
			{
				ID: IdeaRewritePreview, Type: ScreenPreview, Title: "Rewritten Idea", FromStep: IdeaRewriteExec,
				Commands: []CommandDefinition{
					{ID: "/continue", Destination: StepDestination(IdeaReviewExec)},
					{ID: "/edit", Destination: StepDestination(IdeaEdit)},
					{ID: "/cancel", Destination: ScreenDestination(ChangeDetailsTerminal)},
				},
			},
			{
				ID: IdeaReviewPreview, Type: ScreenPreview, Title: "Reviewed Idea", FromStep: IdeaReviewExec,
				Commands: []CommandDefinition{
					{ID: "/continue", Destination: ScreenDestination(MainTerminal)},
					{ID: "/edit", Destination: StepDestination(IdeaEdit)},
					{ID: "/cancel", Destination: ScreenDestination(ChangeDetailsTerminal)},
				},
			},
		},
	}
}
