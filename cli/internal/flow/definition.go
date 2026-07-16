// Package flow provides the reusable, uncomposed CLI Flow runtime.
package flow

// DefinitionID identifies a Flow definition.
type DefinitionID string

// StepID identifies a Step inside a Flow definition.
type StepID string

// TaskID identifies a task inside a Flow definition.
type TaskID string

// ScreenID identifies a reusable Flow or terminal navigation Screen.
type ScreenID string

// CommandID identifies a user-facing Screen command.
type CommandID string

// Artifact identifies the value operated on by a Step.
type Artifact string

const (
	// ArtifactIdea is a Change idea.
	ArtifactIdea Artifact = "idea"
	// ArtifactSpec is a Change spec.
	ArtifactSpec Artifact = "spec"
	// ArtifactPR is a Change pull request body.
	ArtifactPR Artifact = "pr"
	// ArtifactImplement is implementation work.
	ArtifactImplement Artifact = "implement"
	// ArtifactReview is review work.
	ArtifactReview Artifact = "review"
	// ArtifactFinalize is finalization work.
	ArtifactFinalize Artifact = "finalize"
)

// TaskType identifies the operation performed by a task.
type TaskType string

const (
	// TaskEditor opens an artifact in an external editor.
	TaskEditor TaskType = "editor"
	// TaskExec runs a configured non-interactive command.
	TaskExec TaskType = "exec"
	// TaskInteractive resumes a configured interactive session.
	TaskInteractive TaskType = "interactive"
)

// ScreenType identifies a reusable Flow Screen.
type ScreenType string

const (
	// ScreenEditor represents editor activity.
	ScreenEditor ScreenType = "editor"
	// ScreenExec represents a running non-interactive command.
	ScreenExec ScreenType = "exec"
	// ScreenInteractive represents an interactive checkpoint.
	ScreenInteractive ScreenType = "interactive"
	// ScreenPreview represents an artifact preview or diff.
	ScreenPreview ScreenType = "preview"
	// ScreenError represents a concrete runtime error.
	ScreenError ScreenType = "error"
)

// DestinationKind identifies the meaning of a command destination.
type DestinationKind string

const (
	// DestinationStep starts another Step in the current Flow.
	DestinationStep DestinationKind = "step"
	// DestinationScreen navigates to a composition-approved terminal Screen.
	DestinationScreen DestinationKind = "screen"
)

// Destination is a typed reference to either a Step or terminal Screen.
type Destination struct {
	Kind   DestinationKind `yaml:"kind"`
	Step   StepID          `yaml:"step,omitempty"`
	Screen ScreenID        `yaml:"screen,omitempty"`
}

// CommandDefinition maps a user-facing command to a typed destination.
type CommandDefinition struct {
	ID          CommandID   `yaml:"id"`
	Destination Destination `yaml:"destination"`
}

// ScreenDefinition configures one reusable Screen.
type ScreenDefinition struct {
	ID       ScreenID            `yaml:"id"`
	Type     ScreenType          `yaml:"type"`
	Title    string              `yaml:"title,omitempty"`
	Commands []CommandDefinition `yaml:"commands,omitempty"`
	Options  ScreenOptions       `yaml:"options,omitempty"`
}

// ScreenOptions contains typed static presentation options.
type ScreenOptions struct {
	Theme string `yaml:"theme,omitempty"`
}

// TaskDefinition configures one operation and its destinations.
type TaskDefinition struct {
	ID               TaskID   `yaml:"id"`
	Type             TaskType `yaml:"type"`
	Artifact         Artifact `yaml:"artifact"`
	Screen           ScreenID `yaml:"screen"`
	Prompt           string   `yaml:"prompt,omitempty"`
	Script           string   `yaml:"script,omitempty"`
	ExpectedOutput   string   `yaml:"expected_output,omitempty"`
	Preview          ScreenID `yaml:"preview"`
	UnexpectedOutput ScreenID `yaml:"unexpected_output,omitempty"`
	Editor           ScreenID `yaml:"editor,omitempty"`
	Error            ScreenID `yaml:"error"`
}

// StepDefinition configures the ordered tasks for one artifact Step.
type StepDefinition struct {
	ID    StepID           `yaml:"id"`
	Tasks []TaskDefinition `yaml:"tasks"`
}

// Definition is immutable static behavior supplied to the runtime.
type Definition struct {
	ID      DefinitionID       `yaml:"id"`
	Steps   []StepDefinition   `yaml:"steps"`
	Screens []ScreenDefinition `yaml:"screens"`
}

// StepDestination constructs a Step destination.
func StepDestination(step StepID) Destination {
	return Destination{Kind: DestinationStep, Step: step}
}

// ScreenDestination constructs a terminal Screen destination.
func ScreenDestination(screen ScreenID) Destination {
	return Destination{Kind: DestinationScreen, Screen: screen}
}
