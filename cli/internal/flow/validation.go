package flow

import (
	"fmt"
	"strings"
)

// ValidateDefinitions validates unique definitions and their contents.
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

// ValidateDefinition validates a completed definition before execution.
func ValidateDefinition(definition Definition, terminalScreens []ScreenID) error {
	if strings.TrimSpace(string(definition.ID)) == "" {
		return fmt.Errorf("definition.id is required")
	}
	steps, tasks, err := validateIdentifiers(definition)
	if err != nil {
		return err
	}
	screens, err := validateScreens(definition)
	if err != nil {
		return err
	}
	terminals, err := terminalSet(terminalScreens, screens)
	if err != nil {
		return err
	}
	for stepIndex, step := range definition.Steps {
		if err := validateStep(stepIndex, step, steps, screens, terminals); err != nil {
			return err
		}
	}
	for screenIndex, screen := range definition.Screens {
		if err := validateScreen(screenIndex, screen, steps, tasks, screens, terminals); err != nil {
			return err
		}
	}
	return validateTaskPreviewOwnership(definition, screens)
}

func validateIdentifiers(definition Definition) (map[StepID]StepDefinition, map[TaskID]TaskDefinition, error) {
	if len(definition.Steps) == 0 {
		return nil, nil, fmt.Errorf("definition.steps is required")
	}
	steps := make(map[StepID]StepDefinition, len(definition.Steps))
	tasks := make(map[TaskID]TaskDefinition)
	for stepIndex, step := range definition.Steps {
		if strings.TrimSpace(string(step.ID)) == "" {
			return nil, nil, fmt.Errorf("steps[%d].id is required", stepIndex)
		}
		if _, exists := steps[step.ID]; exists {
			return nil, nil, fmt.Errorf("steps[%d].id %q is duplicate", stepIndex, step.ID)
		}
		if len(step.Tasks) == 0 {
			return nil, nil, fmt.Errorf("steps[%d].tasks is required", stepIndex)
		}
		steps[step.ID] = step
		for taskIndex, task := range step.Tasks {
			if strings.TrimSpace(string(task.ID)) == "" {
				return nil, nil, fmt.Errorf("steps[%d].tasks[%d].id is required", stepIndex, taskIndex)
			}
			if _, exists := tasks[task.ID]; exists {
				return nil, nil, fmt.Errorf("steps[%d].tasks[%d].id %q is duplicate", stepIndex, taskIndex, task.ID)
			}
			tasks[task.ID] = task
		}
	}
	return steps, tasks, nil
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
		if screen.Type != ScreenPreview && screen.Options != (ScreenOptions{}) {
			return nil, fmt.Errorf("screens[%d].options is forbidden for %s Screens", index, screen.Type)
		}
		screens[screen.ID] = screen
	}
	return screens, nil
}

func terminalSet(values []ScreenID, runtime map[ScreenID]ScreenDefinition) (map[ScreenID]struct{}, error) {
	terminals := make(map[ScreenID]struct{}, len(values))
	for index, screen := range values {
		if strings.TrimSpace(string(screen)) == "" {
			return nil, fmt.Errorf("terminal_screens[%d] is required", index)
		}
		if _, exists := runtime[screen]; exists {
			return nil, fmt.Errorf("terminal_screens[%d] %q references a runtime Flow Screen", index, screen)
		}
		terminals[screen] = struct{}{}
	}
	return terminals, nil
}

func validateStep(index int, step StepDefinition, steps map[StepID]StepDefinition, screens map[ScreenID]ScreenDefinition, terminals map[ScreenID]struct{}) error {
	if !supportedMode(step.Mode) {
		return fmt.Errorf("steps[%d].mode %q is unsupported", index, step.Mode)
	}
	if !validModeShape(step.Mode, step.Tasks) {
		return fmt.Errorf("steps[%d].tasks do not match %s Mode", index, step.Mode)
	}
	artifact := step.Tasks[0].Artifact
	for taskIndex, task := range step.Tasks {
		field := fmt.Sprintf("steps[%d].tasks[%d]", index, taskIndex)
		if !supportedArtifact(task.Artifact) {
			return fmt.Errorf("%s.artifact %q is unsupported", field, task.Artifact)
		}
		if task.Artifact != artifact {
			return fmt.Errorf("%s.artifact %q does not match Step artifact %q", field, task.Artifact, artifact)
		}
		if err := requireScreenType(field+".screen", task.Screen, taskScreenType(task.Type), screens); err != nil {
			return err
		}
		if err := requireScreenType(field+".preview", task.Preview, ScreenPreview, screens); err != nil {
			return err
		}
		if err := requireScreenType(field+".error", task.Error, ScreenError, screens); err != nil {
			return err
		}
		switch task.Type {
		case TaskEditor:
			if err := validateEditorTask(field, task, screens, terminals); err != nil {
				return err
			}
		case TaskExec:
			if err := validateExecTask(field, task, step, taskIndex, screens); err != nil {
				return err
			}
		case TaskChat:
			if err := validateChatTask(field, task, step, steps, screens); err != nil {
				return err
			}
		default:
			return fmt.Errorf("%s.type %q is unsupported", field, task.Type)
		}
	}
	return nil
}

func validateEditorTask(field string, task TaskDefinition, screens map[ScreenID]ScreenDefinition, terminals map[ScreenID]struct{}) error {
	if task.Prompt != "" || task.Script != "" || task.ExpectedOutput != "" || task.UnexpectedOutput != "" || task.Editor != (Destination{}) {
		return fmt.Errorf("%s contains fields forbidden for editor tasks", field)
	}
	if err := validateTerminalDestination(field+".cancel", task.Cancel, terminals); err != nil {
		return err
	}
	if task.Cancel.Screen != ChangeDetailsTerminal {
		return fmt.Errorf("%s.cancel must reference %q", field, ChangeDetailsTerminal)
	}
	return nil
}

func validateExecTask(field string, task TaskDefinition, step StepDefinition, taskIndex int, screens map[ScreenID]ScreenDefinition) error {
	if strings.TrimSpace(task.Prompt) == "" {
		return fmt.Errorf("%s.prompt is required for exec tasks", field)
	}
	if strings.TrimSpace(task.Script) == "" {
		return fmt.Errorf("%s.script is required for exec tasks", field)
	}
	if task.ExpectedOutput == "" {
		return fmt.Errorf("%s.expected_output is required for exec tasks", field)
	}
	if task.Editor != (Destination{}) || task.Cancel != (Destination{}) {
		return fmt.Errorf("%s contains fields forbidden for exec tasks", field)
	}
	if err := requireScreenType(field+".unexpected_output", task.UnexpectedOutput, ScreenChat, screens); err != nil {
		return err
	}
	if taskIndex+1 >= len(step.Tasks) || step.Tasks[taskIndex+1].Type != TaskChat || step.Tasks[taskIndex+1].Screen != task.UnexpectedOutput || step.Tasks[taskIndex+1].Artifact != task.Artifact {
		return fmt.Errorf("%s.unexpected_output must reference the following same-artifact Chat task", field)
	}
	return nil
}

func validateChatTask(field string, task TaskDefinition, owner StepDefinition, steps map[StepID]StepDefinition, screens map[ScreenID]ScreenDefinition) error {
	if strings.TrimSpace(task.Script) == "" {
		return fmt.Errorf("%s.script is required for chat tasks", field)
	}
	if task.Prompt != "" || task.ExpectedOutput != "" || task.UnexpectedOutput != "" || task.Cancel != (Destination{}) {
		return fmt.Errorf("%s contains fields forbidden for chat tasks", field)
	}
	if task.Editor == (Destination{}) {
		return fmt.Errorf("%s.editor is required for chat tasks", field)
	}
	switch task.Editor.Kind {
	case DestinationStep:
		if task.Editor.Screen != "" {
			return fmt.Errorf("%s.editor.screen is forbidden for step destination", field)
		}
		step, exists := steps[task.Editor.Step]
		if !exists {
			return fmt.Errorf("%s.editor.step references unknown Step %q", field, task.Editor.Step)
		}
		if step.Mode != ModeEditor || len(step.Tasks) != 1 || step.Tasks[0].Artifact != task.Artifact {
			return fmt.Errorf("%s.editor.step must reference a same-artifact Editor Step", field)
		}
	case DestinationScreen:
		if task.Editor.Step != "" {
			return fmt.Errorf("%s.editor.step is forbidden for screen destination", field)
		}
		if err := requireScreenType(field+".editor.screen", task.Editor.Screen, ScreenEditor, screens); err != nil {
			return err
		}
		if !hasEditorTask(task.Editor.Screen, task.Artifact, steps) {
			return fmt.Errorf("%s.editor.screen must reference a same-artifact Editor task", field)
		}
	default:
		return fmt.Errorf("%s.editor.kind %q is unsupported", field, task.Editor.Kind)
	}
	return nil
}

func hasEditorTask(screen ScreenID, artifact Artifact, steps map[StepID]StepDefinition) bool {
	for _, step := range steps {
		if step.Mode != ModeEditor || !validModeShape(step.Mode, step.Tasks) {
			continue
		}
		task := step.Tasks[0]
		if task.Screen == screen && task.Artifact == artifact {
			return true
		}
	}
	return false
}

func validateScreen(index int, screen ScreenDefinition, steps map[StepID]StepDefinition, tasks map[TaskID]TaskDefinition, runtime map[ScreenID]ScreenDefinition, terminals map[ScreenID]struct{}) error {
	_ = tasks
	if screen.Type != ScreenPreview {
		if screen.FromStep != "" {
			return fmt.Errorf("screens[%d].from_step is forbidden for %s Screens", index, screen.Type)
		}
		if len(screen.Commands) != 0 {
			return fmt.Errorf("screens[%d].commands is forbidden for %s Screens", index, screen.Type)
		}
		return nil
	}
	owner, exists := steps[screen.FromStep]
	if !exists {
		return fmt.Errorf("screens[%d].from_step references unknown Step %q", index, screen.FromStep)
	}
	seen := make(map[CommandID]struct{}, len(screen.Commands))
	static := make([]CommandID, 0, len(screen.Commands))
	for commandIndex, command := range screen.Commands {
		field := fmt.Sprintf("screens[%d].commands[%d]", index, commandIndex)
		if strings.TrimSpace(string(command.ID)) == "" {
			return fmt.Errorf("%s.id is required", field)
		}
		if _, duplicate := seen[command.ID]; duplicate {
			return fmt.Errorf("%s.id %q is duplicate", field, command.ID)
		}
		seen[command.ID] = struct{}{}
		static = append(static, command.ID)
		switch command.ID {
		case "/continue":
			if err := validateDestination(field+".destination", command.Destination, steps, runtime, terminals); err != nil {
				return err
			}
		case "/edit":
			if command.Destination.Kind != DestinationStep || command.Destination.Step == "" || command.Destination.Screen != "" {
				return fmt.Errorf("%s.destination must reference an Editor Step", field)
			}
			target, ok := steps[command.Destination.Step]
			if !ok || target.Mode != ModeEditor || target.Tasks[0].Artifact != owner.Tasks[0].Artifact {
				return fmt.Errorf("%s.destination must reference a same-artifact Editor Step", field)
			}
		case "/cancel":
			if err := validateTerminalDestination(field+".destination", command.Destination, terminals); err != nil {
				return err
			}
			if command.Destination.Screen != ChangeDetailsTerminal {
				return fmt.Errorf("%s.destination must reference %q", field, ChangeDetailsTerminal)
			}
		case "/chat":
			if owner.Mode != ModeExec && owner.Mode != ModeChat {
				return fmt.Errorf("%s /chat is forbidden for %s Mode Preview", field, owner.Mode)
			}
			chat, ok := chatTask(owner)
			if !ok || command.Destination.Kind != DestinationScreen || command.Destination.Screen != chat.Screen || command.Destination.Step != "" {
				return fmt.Errorf("%s.destination must reference from-step Chat Screen %q", field, chat.Screen)
			}
		default:
			return fmt.Errorf("%s.id %q is unsupported for Preview", field, command.ID)
		}
	}
	expected := []CommandID{"/continue", "/edit", "/cancel"}
	if owner.Mode == ModeExec || owner.Mode == ModeChat {
		if _, declared := seen["/chat"]; declared {
			expected = []CommandID{"/continue", "/chat", "/edit", "/cancel"}
		}
	}
	if !sameCommands(static, expected) {
		return fmt.Errorf("screens[%d].commands must be %v", index, expected)
	}
	return nil
}

func validateTaskPreviewOwnership(definition Definition, screens map[ScreenID]ScreenDefinition) error {
	for stepIndex, step := range definition.Steps {
		for taskIndex, task := range step.Tasks {
			preview := screens[task.Preview]
			if preview.FromStep != step.ID {
				return fmt.Errorf("steps[%d].tasks[%d].preview %q has from_step %q, want %q", stepIndex, taskIndex, task.Preview, preview.FromStep, step.ID)
			}
		}
	}
	return nil
}

func validateDestination(field string, destination Destination, steps map[StepID]StepDefinition, runtime map[ScreenID]ScreenDefinition, terminals map[ScreenID]struct{}) error {
	switch destination.Kind {
	case DestinationStep:
		if destination.Step == "" || destination.Screen != "" {
			return fmt.Errorf("%s must contain only a Step reference", field)
		}
		if _, exists := steps[destination.Step]; !exists {
			return fmt.Errorf("%s.step references unknown Step %q", field, destination.Step)
		}
	case DestinationScreen:
		if destination.Screen == "" || destination.Step != "" {
			return fmt.Errorf("%s must contain only a Screen reference", field)
		}
		if _, exists := runtime[destination.Screen]; exists {
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

func validateTerminalDestination(field string, destination Destination, terminals map[ScreenID]struct{}) error {
	if destination.Kind != DestinationScreen || destination.Screen == "" || destination.Step != "" {
		return fmt.Errorf("%s must reference a terminal Screen", field)
	}
	if _, ok := terminals[destination.Screen]; !ok {
		return fmt.Errorf("%s.screen references unknown terminal Screen %q", field, destination.Screen)
	}
	return nil
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

func validModeShape(mode Mode, tasks []TaskDefinition) bool {
	switch mode {
	case ModeEditor:
		return len(tasks) == 1 && tasks[0].Type == TaskEditor
	case ModeChat:
		return len(tasks) == 1 && tasks[0].Type == TaskChat
	case ModeExec:
		return len(tasks) == 2 && tasks[0].Type == TaskExec && tasks[1].Type == TaskChat
	default:
		return false
	}
}

func chatTask(step StepDefinition) (TaskDefinition, bool) {
	for _, task := range step.Tasks {
		if task.Type == TaskChat {
			return task, true
		}
	}
	return TaskDefinition{}, false
}

func sameCommands(left, right []CommandID) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func supportedArtifact(artifact Artifact) bool {
	switch artifact {
	case ArtifactIdea, ArtifactSpec, ArtifactPR, ArtifactImplement, ArtifactReview, ArtifactFinalize:
		return true
	default:
		return false
	}
}

func supportedMode(mode Mode) bool {
	return mode == ModeEditor || mode == ModeExec || mode == ModeChat
}

func supportedScreenType(screenType ScreenType) bool {
	return screenType == ScreenEditor || screenType == ScreenExec || screenType == ScreenChat || screenType == ScreenPreview || screenType == ScreenError
}

func taskScreenType(taskType TaskType) ScreenType {
	switch taskType {
	case TaskEditor:
		return ScreenEditor
	case TaskExec:
		return ScreenExec
	case TaskChat:
		return ScreenChat
	default:
		return ""
	}
}
