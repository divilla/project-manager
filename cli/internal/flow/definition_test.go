package flow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ideaTerminals = []ScreenID{MainTerminal, ChangesListTerminal, ChangeDetailsTerminal}

func TestIdeaDefinitionIsFreshAndValid(t *testing.T) {
	first := IdeaDefinition()
	second := IdeaDefinition()
	require.NoError(t, ValidateDefinition(first, ideaTerminals))
	first.Steps[0].Tasks[0].Artifact = ArtifactPR
	assert.Equal(t, ArtifactIdea, second.Steps[0].Tasks[0].Artifact)
	require.NoError(t, ValidateDefinition(second, ideaTerminals))
}

func TestConformanceDefinitionCoversGenericArtifactsAndModes(t *testing.T) {
	definition := ConformanceDefinition()
	require.NoError(t, ValidateDefinition(definition, ideaTerminals))
	assert.Len(t, definition.Steps, 5)
	assert.Equal(t, []Mode{ModeEditor, ModeEditor, ModeExec, ModeEditor, ModeChat}, []Mode{
		definition.Steps[0].Mode,
		definition.Steps[1].Mode,
		definition.Steps[2].Mode,
		definition.Steps[3].Mode,
		definition.Steps[4].Mode,
	})
	assert.Equal(t, ScreenEditor, definition.Screens[2].Type)
	assert.Equal(t, ScreenChat, definition.Screens[4].Type)
	assert.Equal(t, DestinationScreen, definition.Steps[2].Tasks[1].Editor.Kind)
}

func TestValidateDefinitionRejectsCrossArtifactChatEditorScreen(t *testing.T) {
	definition := ConformanceDefinition()
	definition.Steps[4].Tasks[0].Editor = ScreenDestination(definition.Steps[0].Tasks[0].Screen)

	err := ValidateDefinition(definition, ideaTerminals)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "editor.screen must reference a same-artifact Editor task")
}

func TestValidateDefinitionEnforcesModeTaskShapesAndPreviewChat(t *testing.T) {
	tests := []struct {
		name string
		edit func(*Definition)
		want string
	}{
		{name: "empty", edit: func(d *Definition) { d.Steps[1].Tasks = nil }, want: "tasks is required"},
		{name: "standalone exec", edit: func(d *Definition) { d.Steps[1].Tasks = d.Steps[1].Tasks[:1] }, want: "do not match exec Mode"},
		{name: "reverse", edit: func(d *Definition) {
			d.Steps[1].Tasks[0], d.Steps[1].Tasks[1] = d.Steps[1].Tasks[1], d.Steps[1].Tasks[0]
		}, want: "do not match exec Mode"},
		{name: "script", edit: func(d *Definition) { d.Steps[1].Mode = ModeScript }, want: "unsupported"},
		{name: "redirect chat", edit: func(d *Definition) {
			d.Screens[7].Commands = []CommandDefinition{
				{ID: "/continue", Destination: StepDestination(IdeaReviewExec)},
				{ID: "/chat", Destination: ScreenDestination(IdeaReviewChat)},
				{ID: "/edit", Destination: StepDestination(IdeaEdit)},
				{ID: "/cancel", Destination: ScreenDestination(ChangeDetailsTerminal)},
			}
		}, want: "from-step Chat Screen"},
		{name: "preview cancel outside details", edit: func(d *Definition) {
			d.Screens[6].Commands[2].Destination = ScreenDestination(MainTerminal)
		}, want: "ChangeDetailsState"},
		{name: "editor cancel outside details", edit: func(d *Definition) {
			d.Steps[0].Tasks[0].Cancel = ScreenDestination(MainTerminal)
		}, want: "ChangeDetailsState"},
		{name: "edit destination contains screen", edit: func(d *Definition) {
			d.Screens[6].Commands[1].Destination.Screen = ChangeDetailsTerminal
		}, want: "Editor Step"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			definition := IdeaDefinition()
			test.edit(&definition)
			err := ValidateDefinition(definition, ideaTerminals)
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.want)
		})
	}
}

func TestValidateDefinitionAcceptsSingleTaskChatMode(t *testing.T) {
	definition := IdeaDefinition()
	definition.Steps = append(definition.Steps, StepDefinition{
		ID: "IdeaStandaloneChat", Mode: ModeChat,
		Tasks: []TaskDefinition{{
			ID: "IdeaStandaloneChatTask", Type: TaskChat, Artifact: ArtifactIdea,
			Screen: "IdeaStandaloneChatScreen", Script: "scripts/codex-resume-session.sh",
			Editor: StepDestination(IdeaEdit), Preview: "IdeaStandaloneChatPreview", Error: GenericErrorScreen,
		}},
	})
	definition.Screens = append(definition.Screens,
		ScreenDefinition{ID: "IdeaStandaloneChatScreen", Type: ScreenChat},
		ScreenDefinition{
			ID: "IdeaStandaloneChatPreview", Type: ScreenPreview, FromStep: "IdeaStandaloneChat",
			Commands: []CommandDefinition{
				{ID: "/continue", Destination: ScreenDestination(MainTerminal)},
				{ID: "/edit", Destination: StepDestination(IdeaEdit)},
				{ID: "/cancel", Destination: ScreenDestination(ChangeDetailsTerminal)},
			},
		},
	)
	require.NoError(t, ValidateDefinition(definition, ideaTerminals))
}
