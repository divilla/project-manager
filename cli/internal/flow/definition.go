// Package flow provides the reusable CLI Flow runtime.
package flow

// DefinitionID identifies a Flow definition.
type DefinitionID string

// StepID identifies a Step.
type StepID string

// TaskID identifies a Task.
type TaskID string

// ScreenID identifies a runtime or terminal Screen.
type ScreenID string

// CommandID identifies a user-facing command.
type CommandID string

// Artifact identifies content operated on by a Step.
type Artifact string

// Supported artifact identifiers.
const (
	ArtifactIdea      Artifact = "idea"
	ArtifactSpec      Artifact = "spec"
	ArtifactPR        Artifact = "pr"
	ArtifactImplement Artifact = "implement"
	ArtifactReview    Artifact = "review"
	ArtifactFinalize  Artifact = "finalize"
)

// Mode determines the exact supported Task shape of a Step.
type Mode string

// Supported Step modes.
const (
	// ModeEditor owns one Editor Task.
	ModeEditor Mode = "editor"
	ModeExec   Mode = "exec"
	ModeChat   Mode = "chat"
	ModeScript Mode = "script"
)

// TaskType identifies one executable unit in a Step.
type TaskType string

// Supported Task types.
const (
	TaskEditor TaskType = "editor"
	TaskExec   TaskType = "exec"
	TaskChat   TaskType = "chat"
)

// ScreenType identifies one reusable runtime Screen.
type ScreenType string

// Supported Screen types.
const (
	ScreenEditor  ScreenType = "editor"
	ScreenExec    ScreenType = "exec"
	ScreenChat    ScreenType = "chat"
	ScreenPreview ScreenType = "preview"
	ScreenError   ScreenType = "error"
)

// DestinationKind identifies a typed command target.
type DestinationKind string

// Supported destination kinds.
const (
	DestinationStep   DestinationKind = "step"
	DestinationScreen DestinationKind = "screen"
)

// Destination references either a Step or a Screen.
type Destination struct {
	Kind   DestinationKind `yaml:"kind"`
	Step   StepID          `yaml:"step,omitempty"`
	Screen ScreenID        `yaml:"screen,omitempty"`
}

// CommandDefinition maps a command to a typed destination.
type CommandDefinition struct {
	ID          CommandID   `yaml:"id"`
	Destination Destination `yaml:"destination"`
}

// ScreenDefinition configures one reusable Screen.
type ScreenDefinition struct {
	ID       ScreenID            `yaml:"id"`
	Type     ScreenType          `yaml:"type"`
	Title    string              `yaml:"title,omitempty"`
	FromStep StepID              `yaml:"from_step,omitempty"`
	Commands []CommandDefinition `yaml:"commands,omitempty"`
	Options  ScreenOptions       `yaml:"options,omitempty"`
}

// ScreenOptions configures static rendering options.
type ScreenOptions struct {
	Theme string `yaml:"theme,omitempty"`
}

// TaskDefinition configures one ordered Task.
type TaskDefinition struct {
	ID               TaskID      `yaml:"id"`
	Type             TaskType    `yaml:"type"`
	Artifact         Artifact    `yaml:"artifact"`
	Screen           ScreenID    `yaml:"screen"`
	Prompt           string      `yaml:"prompt,omitempty"`
	Script           string      `yaml:"script,omitempty"`
	ExpectedOutput   string      `yaml:"expected_output,omitempty"`
	Preview          ScreenID    `yaml:"preview"`
	UnexpectedOutput ScreenID    `yaml:"unexpected_output,omitempty"`
	Editor           Destination `yaml:"editor,omitempty"`
	Cancel           Destination `yaml:"cancel,omitempty"`
	Error            ScreenID    `yaml:"error"`
}

// StepDefinition owns an ordered collection of Tasks.
type StepDefinition struct {
	ID    StepID           `yaml:"id"`
	Mode  Mode             `yaml:"mode"`
	Tasks []TaskDefinition `yaml:"tasks"`
}

// Definition contains immutable static Flow behavior.
type Definition struct {
	ID      DefinitionID       `yaml:"id"`
	Steps   []StepDefinition   `yaml:"steps"`
	Screens []ScreenDefinition `yaml:"screens"`
}

// StepDestination constructs a typed Step destination.
func StepDestination(step StepID) Destination {
	return Destination{Kind: DestinationStep, Step: step}
}

// ScreenDestination constructs a typed Screen destination.
func ScreenDestination(screen ScreenID) Destination {
	return Destination{Kind: DestinationScreen, Screen: screen}
}
