package flow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestConformanceDefinitionIsFreshYAMLRepresentableAndValid(t *testing.T) {
	first := ConformanceDefinition()
	second := ConformanceDefinition()

	first.Steps[0].ID = "mutated"
	first.Screens[4].Commands[0].ID = "/mutated"
	first.Screens[4].Options.Theme = "changed"

	assert.Equal(t, StepID("idea-edit"), second.Steps[0].ID)
	assert.Equal(t, CommandID("/spec"), second.Screens[4].Commands[0].ID)
	assert.Equal(t, "Coldark-Dark", second.Screens[4].Options.Theme)
	require.NoError(t, ValidateDefinition(second, []ScreenID{"changes"}))

	encoded, err := yaml.Marshal(second)
	require.NoError(t, err)
	assert.NotContains(t, string(encoded), "temp_dir")
	assert.Contains(t, string(encoded), "expected_output:")
	assert.Contains(t, string(encoded), "destination:")

	var decoded Definition
	require.NoError(t, yaml.Unmarshal(encoded, &decoded))
	require.NoError(t, ValidateDefinition(decoded, []ScreenID{"changes"}))
	assert.Equal(t, second, decoded)
}

func TestValidateDefinitionsRejectsDuplicateDefinitionIdentifiers(t *testing.T) {
	definition := ConformanceDefinition()
	err := ValidateDefinitions([]Definition{definition, definition}, []ScreenID{"changes"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "definitions[1].id")
}

func TestValidateDefinitionAcceptsExecOnlyWithReusableInteractiveConfiguration(t *testing.T) {
	definition := ConformanceDefinition()
	interactive := definition.Steps[1].Tasks[1]
	definition.Steps[1].Tasks = definition.Steps[1].Tasks[:1]
	definition.Steps = append(definition.Steps, StepDefinition{
		ID:    "spec-interactive",
		Tasks: []TaskDefinition{interactive},
	})

	require.NoError(t, ValidateDefinition(definition, []ScreenID{"changes"}))
}

func TestValidateDefinitionRejectsInvalidFieldsAndReferences(t *testing.T) {
	tests := []struct {
		name       string
		mutate     func(*Definition)
		terminals  []ScreenID
		diagnostic string
	}{
		{
			name: "missing definition identifier",
			mutate: func(definition *Definition) {
				definition.ID = ""
			},
			diagnostic: "definition.id",
		},
		{
			name: "duplicate Step identifier",
			mutate: func(definition *Definition) {
				definition.Steps[1].ID = definition.Steps[0].ID
			},
			diagnostic: "steps[1].id",
		},
		{
			name: "duplicate task identifier",
			mutate: func(definition *Definition) {
				definition.Steps[1].Tasks[0].ID = definition.Steps[0].Tasks[0].ID
			},
			diagnostic: "steps[1].tasks[0].id",
		},
		{
			name: "duplicate Screen identifier",
			mutate: func(definition *Definition) {
				definition.Screens[1].ID = definition.Screens[0].ID
			},
			diagnostic: "screens[1].id",
		},
		{
			name: "duplicate command identifier",
			mutate: func(definition *Definition) {
				definition.Screens[4].Commands[1].ID = definition.Screens[4].Commands[0].ID
			},
			diagnostic: "screens[4].commands[1].id",
		},
		{
			name: "unsupported artifact",
			mutate: func(definition *Definition) {
				definition.Steps[0].Tasks[0].Artifact = "unknown"
			},
			diagnostic: "steps[0].tasks[0].artifact",
		},
		{
			name: "unsupported Screen type",
			mutate: func(definition *Definition) {
				definition.Screens[0].Type = "unknown"
			},
			diagnostic: "screens[0].type",
		},
		{
			name: "forbidden Screen option",
			mutate: func(definition *Definition) {
				definition.Screens[1].Options.Theme = "other"
			},
			diagnostic: "screens[1].options",
		},
		{
			name: "unsupported task type",
			mutate: func(definition *Definition) {
				definition.Steps[0].Tasks[0].Type = "unknown"
			},
			diagnostic: "steps[0].tasks[0].type",
		},
		{
			name: "unsupported task sequence",
			mutate: func(definition *Definition) {
				definition.Steps[1].Tasks[0], definition.Steps[1].Tasks[1] = definition.Steps[1].Tasks[1], definition.Steps[1].Tasks[0]
			},
			diagnostic: "steps[1].tasks",
		},
		{
			name: "Exec continuation does not target following Interactive Screen",
			mutate: func(definition *Definition) {
				definition.Screens = append(definition.Screens, ScreenDefinition{ID: "other-interactive", Type: ScreenInteractive})
				definition.Steps[1].Tasks[0].UnexpectedOutput = "other-interactive"
			},
			diagnostic: "steps[1].tasks[0].unexpected_output",
		},
		{
			name: "inconsistent artifact",
			mutate: func(definition *Definition) {
				definition.Steps[1].Tasks[1].Artifact = ArtifactIdea
			},
			diagnostic: "steps[1].tasks[1].artifact",
		},
		{
			name: "missing Editor Preview",
			mutate: func(definition *Definition) {
				definition.Steps[0].Tasks[0].Preview = ""
			},
			diagnostic: "steps[0].tasks[0].preview",
		},
		{
			name: "forbidden Editor prompt",
			mutate: func(definition *Definition) {
				definition.Steps[0].Tasks[0].Prompt = "prompt.md"
			},
			diagnostic: "steps[0].tasks[0].prompt",
		},
		{
			name: "missing Exec expected output",
			mutate: func(definition *Definition) {
				definition.Steps[1].Tasks[0].ExpectedOutput = ""
			},
			diagnostic: "steps[1].tasks[0].expected_output",
		},
		{
			name: "wrong Exec unexpected Screen",
			mutate: func(definition *Definition) {
				definition.Steps[1].Tasks[0].UnexpectedOutput = "error"
			},
			diagnostic: "steps[1].tasks[0].unexpected_output",
		},
		{
			name: "missing Interactive script",
			mutate: func(definition *Definition) {
				definition.Steps[1].Tasks[1].Script = ""
			},
			diagnostic: "steps[1].tasks[1].script",
		},
		{
			name: "forbidden Interactive expected output",
			mutate: func(definition *Definition) {
				definition.Steps[1].Tasks[1].ExpectedOutput = "done"
			},
			diagnostic: "steps[1].tasks[1].expected_output",
		},
		{
			name: "unknown Step destination",
			mutate: func(definition *Definition) {
				definition.Screens[4].Commands[0].Destination.Step = "missing"
			},
			diagnostic: "screens[4].commands[0].destination.step",
		},
		{
			name: "destination with both references",
			mutate: func(definition *Definition) {
				definition.Screens[4].Commands[0].Destination.Screen = "changes"
			},
			diagnostic: "screens[4].commands[0].destination.screen",
		},
		{
			name: "unsupported destination kind",
			mutate: func(definition *Definition) {
				definition.Screens[4].Commands[0].Destination.Kind = "unknown"
			},
			diagnostic: "screens[4].commands[0].destination.kind",
		},
		{
			name: "unknown terminal Screen",
			mutate: func(definition *Definition) {
				definition.Screens[4].Commands[1].Destination.Screen = "other"
			},
			diagnostic: "screens[4].commands[1].destination.screen",
		},
		{
			name: "runtime Screen as terminal destination",
			mutate: func(definition *Definition) {
				definition.Screens[4].Commands[1].Destination.Screen = "error"
			},
			diagnostic: "screens[4].commands[1].destination.screen",
		},
		{
			name: "missing command destination",
			mutate: func(definition *Definition) {
				definition.Screens[4].Commands[0].Destination = Destination{}
			},
			diagnostic: "screens[4].commands[0].destination.kind",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			definition := ConformanceDefinition()
			test.mutate(&definition)
			terminals := test.terminals
			if terminals == nil {
				terminals = []ScreenID{"changes"}
			}
			err := ValidateDefinition(definition, terminals)
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.diagnostic)
		})
	}
}

func TestComposeRejectsInvalidDefinitionBeforeExecution(t *testing.T) {
	definition := ConformanceDefinition()
	definition.Steps[0].Tasks[0].Artifact = "unsupported"
	store := &fakeStore{}
	operations := &fakeOperations{}

	model := Compose(Composition{
		Definition:      definition,
		Context:         Context{TempDir: t.TempDir(), ChangeID: 1, Origin: "changes", Step: "idea-edit"},
		TerminalScreens: []ScreenID{"changes"},
		Store:           store,
		Operations:      operations,
	})

	require.Error(t, model.Error())
	assert.Contains(t, model.Error().Error(), "steps[0].tasks[0].artifact")
	assert.Nil(t, model.Init())
	assert.Empty(t, store.loads)
	assert.Equal(t, []CommandID{"/return"}, model.Commands())
}
