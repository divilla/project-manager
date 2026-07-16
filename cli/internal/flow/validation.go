package flow

import (
	"fmt"
	"strings"
)

// ValidateDefinitions validates definition identifiers and every contained definition.
func ValidateDefinitions(definitions []Definition, terminalScreens []ScreenID) error {
	seen := make(map[DefinitionID]struct{}, len(definitions))
	for index, definition := range definitions {
		if strings.TrimSpace(string(definition.ID)) == "" {
			return fmt.Errorf("definitions[%d].id is required", index)
		}
		if _, exists := seen[definition.ID]; exists {
			return fmt.Errorf("definitions[%d].id %q is duplicate", index, definition.ID)
		}
		seen[definition.ID] = struct{}{}
		if err := ValidateDefinition(definition, terminalScreens); err != nil {
			return err
		}
	}
	return nil
}

// ValidateDefinition validates a completed Flow definition before execution.
func ValidateDefinition(definition Definition, terminalScreens []ScreenID) error {
	if strings.TrimSpace(string(definition.ID)) == "" {
		return fmt.Errorf("definition.id is required")
	}
	steps, err := validateIdentifiers(definition)
	if err != nil {
		return err
	}
	screens, err := validateScreens(definition)
	if err != nil {
		return err
	}
	terminals := make(map[ScreenID]struct{}, len(terminalScreens))
	for index, screen := range terminalScreens {
		if strings.TrimSpace(string(screen)) == "" {
			return fmt.Errorf("terminal_screens[%d] is required", index)
		}
		if _, runtimeScreen := screens[screen]; runtimeScreen {
			return fmt.Errorf("terminal_screens[%d] %q references a runtime Flow Screen", index, screen)
		}
		terminals[screen] = struct{}{}
	}
	for stepIndex, step := range definition.Steps {
		if err := validateStep(stepIndex, step, screens); err != nil {
			return err
		}
	}
	if err := validateExecInteractiveConfiguration(definition); err != nil {
		return err
	}
	for screenIndex, screen := range definition.Screens {
		if err := validateCommands(screenIndex, screen, steps, screens, terminals); err != nil {
			return err
		}
	}
	return nil
}

func validateExecInteractiveConfiguration(definition Definition) error {
	configured := make(map[string]struct{})
	for _, step := range definition.Steps {
		for _, task := range step.Tasks {
			if task.Type == TaskInteractive {
				configured[interactiveConfigurationKey(task.Artifact, task.Screen)] = struct{}{}
			}
		}
	}
	for stepIndex, step := range definition.Steps {
		for taskIndex, task := range step.Tasks {
			if task.Type != TaskExec {
				continue
			}
			if _, exists := configured[interactiveConfigurationKey(task.Artifact, task.UnexpectedOutput)]; !exists {
				return fmt.Errorf("steps[%d].tasks[%d].unexpected_output %q has no Interactive task configuration for artifact %q", stepIndex, taskIndex, task.UnexpectedOutput, task.Artifact)
			}
		}
	}
	return nil
}

func interactiveConfigurationKey(artifact Artifact, screen ScreenID) string {
	return string(artifact) + "\x00" + string(screen)
}

func validateIdentifiers(definition Definition) (map[StepID]struct{}, error) {
	steps := make(map[StepID]struct{}, len(definition.Steps))
	tasks := make(map[TaskID]struct{})
	if len(definition.Steps) == 0 {
		return nil, fmt.Errorf("definition.steps is required")
	}
	for stepIndex, step := range definition.Steps {
		if strings.TrimSpace(string(step.ID)) == "" {
			return nil, fmt.Errorf("steps[%d].id is required", stepIndex)
		}
		if _, exists := steps[step.ID]; exists {
			return nil, fmt.Errorf("steps[%d].id %q is duplicate", stepIndex, step.ID)
		}
		steps[step.ID] = struct{}{}
		if len(step.Tasks) == 0 {
			return nil, fmt.Errorf("steps[%d].tasks is required", stepIndex)
		}
		for taskIndex, task := range step.Tasks {
			if strings.TrimSpace(string(task.ID)) == "" {
				return nil, fmt.Errorf("steps[%d].tasks[%d].id is required", stepIndex, taskIndex)
			}
			if _, exists := tasks[task.ID]; exists {
				return nil, fmt.Errorf("steps[%d].tasks[%d].id %q is duplicate", stepIndex, taskIndex, task.ID)
			}
			tasks[task.ID] = struct{}{}
		}
	}
	return steps, nil
}

func validateScreens(definition Definition) (map[ScreenID]ScreenDefinition, error) {
	if len(definition.Screens) == 0 {
		return nil, fmt.Errorf("definition.screens is required")
	}
	screens := make(map[ScreenID]ScreenDefinition, len(definition.Screens))
	for index, screen := range definition.Screens {
		if strings.TrimSpace(string(screen.ID)) == "" {
			return nil, fmt.Errorf("screens[%d].id is required", index)
		}
		if _, exists := screens[screen.ID]; exists {
			return nil, fmt.Errorf("screens[%d].id %q is duplicate", index, screen.ID)
		}
		if !supportedScreenType(screen.Type) {
			return nil, fmt.Errorf("screens[%d].type %q is unsupported", index, screen.Type)
		}
		if err := validateScreenOptions(index, screen); err != nil {
			return nil, err
		}
		screens[screen.ID] = screen
	}
	return screens, nil
}

func validateScreenOptions(index int, screen ScreenDefinition) error {
	switch screen.Type {
	case ScreenPreview:
		return nil
	default:
		if screen.Options != (ScreenOptions{}) {
			return fmt.Errorf("screens[%d].options is forbidden for %s Screens", index, screen.Type)
		}
	}
	return nil
}

func validateStep(stepIndex int, step StepDefinition, screens map[ScreenID]ScreenDefinition) error {
	for taskIndex, task := range step.Tasks {
		if !supportedTaskType(task.Type) {
			return fmt.Errorf("steps[%d].tasks[%d].type %q is unsupported", stepIndex, taskIndex, task.Type)
		}
	}
	if !supportedTaskSequence(step.Tasks) {
		return fmt.Errorf("steps[%d].tasks has unsupported task sequence", stepIndex)
	}
	if len(step.Tasks) == 2 && step.Tasks[0].UnexpectedOutput != step.Tasks[1].Screen {
		return fmt.Errorf("steps[%d].tasks[0].unexpected_output %q does not reference the following Interactive Screen %q", stepIndex, step.Tasks[0].UnexpectedOutput, step.Tasks[1].Screen)
	}
	artifact := step.Tasks[0].Artifact
	for taskIndex, task := range step.Tasks {
		field := fmt.Sprintf("steps[%d].tasks[%d]", stepIndex, taskIndex)
		if !supportedArtifact(task.Artifact) {
			return fmt.Errorf("%s.artifact %q is unsupported", field, task.Artifact)
		}
		if task.Artifact != artifact {
			return fmt.Errorf("%s.artifact %q does not match Step artifact %q", field, task.Artifact, artifact)
		}
		if err := requireScreenType(field+".screen", task.Screen, taskScreenType(task.Type), screens); err != nil {
			return err
		}
		switch task.Type {
		case TaskEditor:
			if err := validateEditorTask(field, task, screens); err != nil {
				return err
			}
		case TaskExec:
			if err := validateExecTask(field, task, screens); err != nil {
				return err
			}
		case TaskInteractive:
			if err := validateInteractiveTask(field, task, screens); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateEditorTask(field string, task TaskDefinition, screens map[ScreenID]ScreenDefinition) error {
	if err := requiredRuntimeDestinations(field, task, screens); err != nil {
		return err
	}
	if task.Prompt != "" {
		return fmt.Errorf("%s.prompt is forbidden for editor tasks", field)
	}
	if task.Script != "" {
		return fmt.Errorf("%s.script is forbidden for editor tasks", field)
	}
	if task.ExpectedOutput != "" {
		return fmt.Errorf("%s.expected_output is forbidden for editor tasks", field)
	}
	if task.UnexpectedOutput != "" {
		return fmt.Errorf("%s.unexpected_output is forbidden for editor tasks", field)
	}
	if task.Editor != "" {
		return fmt.Errorf("%s.editor is forbidden for editor tasks", field)
	}
	return nil
}

func validateExecTask(field string, task TaskDefinition, screens map[ScreenID]ScreenDefinition) error {
	if err := requiredRuntimeDestinations(field, task, screens); err != nil {
		return err
	}
	if strings.TrimSpace(task.Prompt) == "" {
		return fmt.Errorf("%s.prompt is required for exec tasks", field)
	}
	if strings.TrimSpace(task.Script) == "" {
		return fmt.Errorf("%s.script is required for exec tasks", field)
	}
	if task.ExpectedOutput == "" {
		return fmt.Errorf("%s.expected_output is required for exec tasks", field)
	}
	if err := requireScreenType(field+".unexpected_output", task.UnexpectedOutput, ScreenInteractive, screens); err != nil {
		return err
	}
	if task.Editor != "" {
		return fmt.Errorf("%s.editor is forbidden for exec tasks", field)
	}
	return nil
}

func validateInteractiveTask(field string, task TaskDefinition, screens map[ScreenID]ScreenDefinition) error {
	if err := requiredRuntimeDestinations(field, task, screens); err != nil {
		return err
	}
	if strings.TrimSpace(task.Script) == "" {
		return fmt.Errorf("%s.script is required for interactive tasks", field)
	}
	if err := requireScreenType(field+".editor", task.Editor, ScreenEditor, screens); err != nil {
		return err
	}
	if task.Prompt != "" {
		return fmt.Errorf("%s.prompt is forbidden for interactive tasks", field)
	}
	if task.ExpectedOutput != "" {
		return fmt.Errorf("%s.expected_output is forbidden for interactive tasks", field)
	}
	if task.UnexpectedOutput != "" {
		return fmt.Errorf("%s.unexpected_output is forbidden for interactive tasks", field)
	}
	return nil
}

func requiredRuntimeDestinations(field string, task TaskDefinition, screens map[ScreenID]ScreenDefinition) error {
	if err := requireScreenType(field+".preview", task.Preview, ScreenPreview, screens); err != nil {
		return err
	}
	return requireScreenType(field+".error", task.Error, ScreenError, screens)
}

func requireScreenType(field string, id ScreenID, expected ScreenType, screens map[ScreenID]ScreenDefinition) error {
	if strings.TrimSpace(string(id)) == "" {
		return fmt.Errorf("%s is required", field)
	}
	screen, exists := screens[id]
	if !exists {
		return fmt.Errorf("%s references unknown Screen %q", field, id)
	}
	if screen.Type != expected {
		return fmt.Errorf("%s references %s Screen %q, want %s", field, screen.Type, id, expected)
	}
	return nil
}

func validateCommands(screenIndex int, screen ScreenDefinition, steps map[StepID]struct{}, runtimeScreens map[ScreenID]ScreenDefinition, terminals map[ScreenID]struct{}) error {
	if screen.Type != ScreenPreview && len(screen.Commands) > 0 {
		return fmt.Errorf("screens[%d].commands is forbidden for %s Screens", screenIndex, screen.Type)
	}
	seen := make(map[CommandID]struct{}, len(screen.Commands))
	for commandIndex, command := range screen.Commands {
		field := fmt.Sprintf("screens[%d].commands[%d]", screenIndex, commandIndex)
		if strings.TrimSpace(string(command.ID)) == "" {
			return fmt.Errorf("%s.id is required", field)
		}
		if _, exists := seen[command.ID]; exists {
			return fmt.Errorf("%s.id %q is duplicate", field, command.ID)
		}
		seen[command.ID] = struct{}{}
		if err := validateDestination(field+".destination", command.Destination, steps, runtimeScreens, terminals); err != nil {
			return err
		}
	}
	return nil
}

func validateDestination(field string, destination Destination, steps map[StepID]struct{}, runtimeScreens map[ScreenID]ScreenDefinition, terminals map[ScreenID]struct{}) error {
	switch destination.Kind {
	case DestinationStep:
		if destination.Step == "" {
			return fmt.Errorf("%s.step is required for step destination", field)
		}
		if destination.Screen != "" {
			return fmt.Errorf("%s.screen is forbidden for step destination", field)
		}
		if _, exists := steps[destination.Step]; !exists {
			return fmt.Errorf("%s.step references unknown Step %q", field, destination.Step)
		}
	case DestinationScreen:
		if destination.Screen == "" {
			return fmt.Errorf("%s.screen is required for screen destination", field)
		}
		if destination.Step != "" {
			return fmt.Errorf("%s.step is forbidden for screen destination", field)
		}
		if _, exists := runtimeScreens[destination.Screen]; exists {
			return fmt.Errorf("%s.screen %q is a runtime Flow Screen", field, destination.Screen)
		}
		if _, exists := terminals[destination.Screen]; !exists {
			return fmt.Errorf("%s.screen references unknown terminal Screen %q", field, destination.Screen)
		}
	default:
		return fmt.Errorf("%s.kind %q is unsupported", field, destination.Kind)
	}
	return nil
}

func supportedTaskSequence(tasks []TaskDefinition) bool {
	if len(tasks) == 1 {
		return tasks[0].Type == TaskEditor || tasks[0].Type == TaskExec || tasks[0].Type == TaskInteractive
	}
	return len(tasks) == 2 && tasks[0].Type == TaskExec && tasks[1].Type == TaskInteractive
}

func supportedArtifact(artifact Artifact) bool {
	switch artifact {
	case ArtifactIdea, ArtifactSpec, ArtifactPR, ArtifactImplement, ArtifactReview, ArtifactFinalize:
		return true
	default:
		return false
	}
}

func supportedTaskType(taskType TaskType) bool {
	return taskType == TaskEditor || taskType == TaskExec || taskType == TaskInteractive
}

func supportedScreenType(screenType ScreenType) bool {
	switch screenType {
	case ScreenEditor, ScreenExec, ScreenInteractive, ScreenPreview, ScreenError:
		return true
	default:
		return false
	}
}

func taskScreenType(taskType TaskType) ScreenType {
	switch taskType {
	case TaskEditor:
		return ScreenEditor
	case TaskExec:
		return ScreenExec
	case TaskInteractive:
		return ScreenInteractive
	default:
		return ""
	}
}
