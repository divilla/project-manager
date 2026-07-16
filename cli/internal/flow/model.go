package flow

import (
	"context"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Context contains runtime-only values supplied by Flow composition.
type Context struct {
	TempDir         string
	FlowDir         string
	ChangeID        int
	ChangeRef       string
	Origin          ScreenID
	Step            StepID
	Task            TaskID
	Artifact        Artifact
	SessionID       string
	ExecutionResult string
}

// Composition supplies a completed definition and all runtime boundaries.
type Composition struct {
	Definition      Definition
	Context         Context
	TerminalScreens []ScreenID
	Store           ArtifactStore
	Operations      Operations
}

// CommandMsg selects a user-facing command on the current Flow Screen.
type CommandMsg struct {
	ID CommandID
}

// PreviewMode identifies artifact or diff rendering.
type PreviewMode string

const (
	// PreviewArtifact renders output.md.
	PreviewArtifact PreviewMode = "preview"
	// PreviewDiff renders input.md against output.md.
	PreviewDiff PreviewMode = "diff"
)

// Model is the reusable Bubble Tea Flow runtime model.
type Model struct {
	definition Definition
	context    Context
	store      ArtifactStore
	operations Operations
	steps      map[StepID]StepDefinition
	screens    map[ScreenID]ScreenDefinition

	screen         ScreenID
	taskIndex      int
	previewMode    PreviewMode
	rendered       string
	baseline       []byte
	err            error
	running        bool
	done           bool
	terminalScreen ScreenID
	cancel         context.CancelFunc
}

type stepLoadedMsg struct {
	stepID   StepID
	artifact Artifact
	baseline []byte
	err      error
}

type editorFinishedMsg struct {
	err error
}

type editorPreparedMsg struct {
	err error
}

type execFinishedMsg struct {
	err error
}

type execEvaluatedMsg struct {
	output    string
	finalLine string
	err       error
}

type interactivePreparedMsg struct {
	sessionID string
	err       error
}

type interactiveFinishedMsg struct {
	err error
}

type interactiveOutputMsg struct {
	output string
	err    error
}

type stepCompletedMsg struct {
	err error
}

type renderedMsg struct {
	mode   PreviewMode
	result RenderResult
	err    error
}

// Compose validates and prepares an unstarted Flow runtime.
func Compose(composition Composition) Model {
	definition := cloneDefinition(composition.Definition)
	model := Model{
		definition:  definition,
		context:     composition.Context,
		store:       composition.Store,
		operations:  composition.Operations,
		steps:       make(map[StepID]StepDefinition, len(definition.Steps)),
		screens:     make(map[ScreenID]ScreenDefinition, len(definition.Screens)),
		previewMode: PreviewArtifact,
	}
	for _, step := range definition.Steps {
		model.steps[step.ID] = step
	}
	for _, screen := range definition.Screens {
		model.screens[screen.ID] = screen
	}
	if err := ValidateDefinition(definition, composition.TerminalScreens); err != nil {
		return model.withError(fmt.Errorf("invalid Flow definition: %w", err), "")
	}
	if model.store == nil {
		return model.withError(fmt.Errorf("artifact store is required"), "")
	}
	if model.operations == nil {
		return model.withError(fmt.Errorf("external operations boundary is required"), "")
	}
	if strings.TrimSpace(string(model.context.Origin)) == "" {
		return model.withError(fmt.Errorf("flow context originating Screen is required"), "")
	}
	originAllowed := false
	for _, terminal := range composition.TerminalScreens {
		if terminal == model.context.Origin {
			originAllowed = true
			break
		}
	}
	if !originAllowed {
		return model.withError(fmt.Errorf("flow context originating Screen %q is not an allowed terminal Screen", model.context.Origin), "")
	}
	return model
}

// Init starts the configured Step through an asynchronous load command.
func (m Model) Init() tea.Cmd {
	if m.err != nil || m.done {
		return nil
	}
	return m.startStepCommand(m.context.Step)
}

// Update applies typed operation results and navigation commands.
func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case stepLoadedMsg:
		return m.updateStepLoaded(message)
	case editorFinishedMsg:
		return m.updateEditorFinished(message)
	case editorPreparedMsg:
		return m.updateEditorPrepared(message)
	case execFinishedMsg:
		return m.updateExecFinished(message)
	case execEvaluatedMsg:
		return m.updateExecEvaluated(message)
	case interactiveOutputMsg:
		if message.err != nil {
			return m.withTaskError(message.err), nil
		}
		m.context.ExecutionResult = message.output
		return m, nil
	case interactivePreparedMsg:
		return m.updateInteractivePrepared(message)
	case interactiveFinishedMsg:
		return m.updateInteractiveFinished(message)
	case stepCompletedMsg:
		return m.updateStepCompleted(message)
	case renderedMsg:
		return m.updateRendered(message)
	case CommandMsg:
		return m.updateCommand(message.ID)
	case tea.KeyMsg:
		if message.String() == "left" || message.String() == "right" {
			return m.togglePreview()
		}
	}
	return m, nil
}

// View renders the current generic Screen from model state only.
func (m Model) View() string {
	if m.done {
		return ""
	}
	screen, exists := m.screens[m.screen]
	title := screen.Title
	if title == "" && exists {
		title = string(screen.Type)
	}
	if m.err != nil {
		return joinView(title, m.err.Error(), "/return")
	}
	switch screen.Type {
	case ScreenExec:
		return joinView(title, "/stop")
	case ScreenInteractive:
		return joinView(title, m.context.ExecutionResult, "/interactive  /edit  /cancel")
	case ScreenPreview:
		commands := m.Commands()
		labels := make([]string, 0, len(commands))
		for _, command := range commands {
			labels = append(labels, string(command))
		}
		return joinView(title, string(m.previewMode), m.rendered, strings.Join(labels, "  "))
	case ScreenEditor:
		return joinView(title)
	case ScreenError:
		return joinView(title, "/return")
	default:
		return ""
	}
}

// Screen returns the active reusable Flow Screen.
func (m Model) Screen() ScreenID {
	return m.screen
}

// FlowContext returns a copy of the current runtime context.
func (m Model) FlowContext() Context {
	return m.context
}

// Commands returns only commands available on the current Screen.
func (m Model) Commands() []CommandID {
	if m.done {
		return nil
	}
	if m.err != nil {
		return []CommandID{"/return"}
	}
	screen := m.screens[m.screen]
	switch screen.Type {
	case ScreenExec:
		if m.running {
			return []CommandID{"/stop"}
		}
	case ScreenInteractive:
		return []CommandID{"/interactive", "/edit", "/cancel"}
	case ScreenPreview:
		commands := make([]CommandID, 0, len(screen.Commands))
		for _, command := range screen.Commands {
			commands = append(commands, command.ID)
		}
		return commands
	}
	return nil
}

// Error returns the current concrete runtime error.
func (m Model) Error() error {
	return m.err
}

// Rendered returns the latest Preview or Diff output.
func (m Model) Rendered() string {
	return m.rendered
}

// Mode returns the current Preview rendering mode.
func (m Model) Mode() PreviewMode {
	return m.previewMode
}

// Done reports whether the Flow navigated to a terminal Screen.
func (m Model) Done() bool {
	return m.done
}

// TerminalScreen returns the selected terminal Screen after Flow completion.
func (m Model) TerminalScreen() ScreenID {
	return m.terminalScreen
}

func (m Model) updateStepLoaded(message stepLoadedMsg) (tea.Model, tea.Cmd) {
	if message.err != nil {
		return m.withError(message.err, ""), nil
	}
	step, exists := m.steps[message.stepID]
	if !exists {
		return m.withError(fmt.Errorf("unknown Step %q", message.stepID), ""), nil
	}
	m.context.Step = message.stepID
	m.context.Artifact = message.artifact
	m.context.ExecutionResult = ""
	m.baseline = append([]byte(nil), message.baseline...)
	m.taskIndex = 0
	m.previewMode = PreviewArtifact
	m.rendered = ""
	return m.enterTask(step.Tasks[0])
}

func (m Model) updateEditorFinished(message editorFinishedMsg) (tea.Model, tea.Cmd) {
	if m.done || m.screens[m.screen].Type != ScreenEditor {
		return m, nil
	}
	if message.err != nil {
		return m.withTaskError(fmt.Errorf("editor failed: %w", message.err)), nil
	}
	return m, m.completeStepCommand()
}

func (m Model) updateEditorPrepared(message editorPreparedMsg) (tea.Model, tea.Cmd) {
	if message.err != nil {
		return m.withTaskError(message.err), nil
	}
	return m, m.launchEditorCommand()
}

func (m Model) updateExecFinished(message execFinishedMsg) (tea.Model, tea.Cmd) {
	if m.done || m.screens[m.screen].Type != ScreenExec {
		return m, nil
	}
	m.running = false
	if m.cancel != nil {
		m.cancel()
	}
	m.cancel = nil
	if message.err != nil {
		return m.withTaskError(fmt.Errorf("execution failed: %w", message.err)), nil
	}
	return m, m.evaluateExecCommand()
}

func (m Model) updateExecEvaluated(message execEvaluatedMsg) (tea.Model, tea.Cmd) {
	if message.err != nil {
		return m.withTaskError(message.err), nil
	}
	m.context.ExecutionResult = message.output
	task, err := m.activeTask()
	if err != nil {
		return m.withTaskError(err), nil
	}
	if message.finalLine == task.ExpectedOutput {
		return m, m.completeStepCommand()
	}
	step := m.steps[m.context.Step]
	if m.taskIndex+1 < len(step.Tasks) && step.Tasks[m.taskIndex+1].Type == TaskInteractive {
		m.taskIndex++
		return m.enterTask(step.Tasks[m.taskIndex])
	}
	m.screen = task.UnexpectedOutput
	return m, nil
}

func (m Model) updateInteractivePrepared(message interactivePreparedMsg) (tea.Model, tea.Cmd) {
	if message.err != nil {
		return m.withTaskError(message.err), nil
	}
	m.context.SessionID = message.sessionID
	task, err := m.activeInteractiveTask()
	if err != nil {
		return m.withTaskError(err), nil
	}
	workspace := Workspace{TempDir: m.context.TempDir, Artifact: m.context.Artifact}
	directory, err := workspace.Directory()
	if err != nil {
		return m.withTaskError(err), nil
	}
	operationContext, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	m.running = true
	request := InteractiveRequest{
		Script:    task.Script,
		FlowDir:   m.context.FlowDir,
		Workspace: directory,
		SessionID: message.sessionID,
		ChangeID:  m.context.ChangeID,
		ChangeRef: m.context.ChangeRef,
		Artifact:  m.context.Artifact,
	}
	return m, m.operations.Interactive(operationContext, request, func(err error) tea.Msg {
		return interactiveFinishedMsg{err: err}
	})
}

func (m Model) updateInteractiveFinished(message interactiveFinishedMsg) (tea.Model, tea.Cmd) {
	if m.done || m.screens[m.screen].Type != ScreenInteractive {
		return m, nil
	}
	m.running = false
	if m.cancel != nil {
		m.cancel()
	}
	m.cancel = nil
	if message.err != nil {
		return m.withTaskError(fmt.Errorf("interactive session failed: %w", message.err)), nil
	}
	return m, m.completeStepCommand()
}

func (m Model) updateStepCompleted(message stepCompletedMsg) (tea.Model, tea.Cmd) {
	if message.err != nil {
		return m.withTaskError(message.err), nil
	}
	task, err := m.activeTask()
	if err != nil {
		return m.withTaskError(err), nil
	}
	m.screen = task.Preview
	m.previewMode = PreviewArtifact
	m.rendered = ""
	return m, m.renderCommand(PreviewArtifact)
}

func (m Model) updateRendered(message renderedMsg) (tea.Model, tea.Cmd) {
	if message.err != nil {
		return m.withTaskError(fmt.Errorf("%s rendering failed: %w", message.mode, message.err)), nil
	}
	if message.mode == PreviewDiff && message.result.Status > 1 {
		return m.withTaskError(fmt.Errorf("diff command failed with status %d", message.result.Status)), nil
	}
	m.previewMode = message.mode
	m.rendered = message.result.Output
	return m, nil
}

func (m Model) updateCommand(command CommandID) (tea.Model, tea.Cmd) {
	if m.done {
		return m, nil
	}
	if m.err != nil {
		if command == "/return" {
			return m.finish(m.context.Origin), nil
		}
		return m, nil
	}
	screen := m.screens[m.screen]
	switch screen.Type {
	case ScreenExec:
		if command == "/stop" && m.running {
			if m.cancel != nil {
				m.cancel()
			}
			m.running = false
			m.cancel = nil
			return m.finish(m.context.Origin), nil
		}
	case ScreenInteractive:
		switch command {
		case "/cancel":
			if m.cancel != nil {
				m.cancel()
			}
			return m.finish(m.context.Origin), nil
		case "/edit":
			task, err := m.activeInteractiveTask()
			if err != nil {
				return m.withTaskError(err), nil
			}
			m.screen = task.Editor
			return m, m.prepareEditorCommand()
		case "/interactive":
			return m, m.prepareInteractiveCommand()
		}
	case ScreenPreview:
		for _, defined := range screen.Commands {
			if defined.ID != command {
				continue
			}
			if defined.Destination.Kind == DestinationStep {
				m.context.Step = defined.Destination.Step
				m.context.Task = ""
				m.context.Artifact = ""
				m.context.ExecutionResult = ""
				return m, m.startStepCommand(defined.Destination.Step)
			}
			return m.finish(defined.Destination.Screen), nil
		}
	}
	return m, nil
}

func (m Model) togglePreview() (tea.Model, tea.Cmd) {
	if m.err != nil || m.done || m.screens[m.screen].Type != ScreenPreview {
		return m, nil
	}
	mode := PreviewArtifact
	if m.previewMode == PreviewArtifact {
		mode = PreviewDiff
	}
	return m, m.renderCommand(mode)
}

func (m Model) enterTask(task TaskDefinition) (tea.Model, tea.Cmd) {
	m.context.Task = task.ID
	m.context.Artifact = task.Artifact
	m.screen = task.Screen
	m.err = nil
	switch task.Type {
	case TaskEditor:
		return m, m.prepareEditorCommand()
	case TaskExec:
		return m.startExec(task)
	case TaskInteractive:
		return m, m.readOptionalAgentOutputCommand()
	default:
		return m.withTaskError(fmt.Errorf("unsupported task type %q", task.Type)), nil
	}
}

func (m Model) startExec(task TaskDefinition) (tea.Model, tea.Cmd) {
	workspace := Workspace{TempDir: m.context.TempDir, Artifact: task.Artifact}
	directory, err := workspace.Directory()
	if err != nil {
		return m.withTaskError(err), nil
	}
	operationContext, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	m.running = true
	request := ExecRequest{
		Script:    task.Script,
		Prompt:    task.Prompt,
		FlowDir:   m.context.FlowDir,
		Workspace: directory,
		ChangeID:  m.context.ChangeID,
		ChangeRef: m.context.ChangeRef,
		Artifact:  task.Artifact,
	}
	return m, m.operations.Exec(operationContext, request, func(err error) tea.Msg {
		return execFinishedMsg{err: err}
	})
}

func (m Model) prepareEditorCommand() tea.Cmd {
	return func() tea.Msg {
		workspace := Workspace{TempDir: m.context.TempDir, Artifact: m.context.Artifact}
		output, err := workspace.OutputPath()
		if err != nil {
			return editorPreparedMsg{err: err}
		}
		if _, err := os.ReadFile(output); err != nil {
			return editorPreparedMsg{err: fmt.Errorf("read output.md before editor: %w", err)}
		}
		return editorPreparedMsg{}
	}
}

func (m Model) launchEditorCommand() tea.Cmd {
	workspace := Workspace{TempDir: m.context.TempDir, Artifact: m.context.Artifact}
	output, err := workspace.OutputPath()
	if err != nil {
		return func() tea.Msg { return editorFinishedMsg{err: err} }
	}
	directory, err := workspace.Directory()
	if err != nil {
		return func() tea.Msg { return editorFinishedMsg{err: err} }
	}
	return m.operations.Editor(output, directory, func(err error) tea.Msg {
		return editorFinishedMsg{err: err}
	})
}

func (m Model) startStepCommand(stepID StepID) tea.Cmd {
	return func() tea.Msg {
		if strings.TrimSpace(string(stepID)) == "" {
			return stepLoadedMsg{err: fmt.Errorf("flow context current Step is required")}
		}
		step, exists := m.steps[stepID]
		if !exists {
			return stepLoadedMsg{stepID: stepID, err: fmt.Errorf("flow context references unknown Step %q", stepID)}
		}
		if m.context.ChangeID <= 0 {
			return stepLoadedMsg{stepID: stepID, err: fmt.Errorf("flow context Change ID must be a valid positive number")}
		}
		artifact := step.Tasks[0].Artifact
		workspace := Workspace{TempDir: m.context.TempDir, Artifact: artifact}
		if _, err := workspace.Directory(); err != nil {
			return stepLoadedMsg{stepID: stepID, artifact: artifact, err: err}
		}
		content, err := m.store.Load(m.context.ChangeID, artifact)
		if err != nil {
			return stepLoadedMsg{stepID: stepID, artifact: artifact, err: fmt.Errorf("load %s artifact: %w", artifact, err)}
		}
		if err := workspace.replaceBaseline(content); err != nil {
			return stepLoadedMsg{stepID: stepID, artifact: artifact, err: err}
		}
		return stepLoadedMsg{stepID: stepID, artifact: artifact, baseline: append([]byte(nil), content...)}
	}
}

func (m Model) evaluateExecCommand() tea.Cmd {
	return func() tea.Msg {
		workspace := Workspace{TempDir: m.context.TempDir, Artifact: m.context.Artifact}
		output, err := workspace.readAgentOutput(true)
		if err != nil {
			return execEvaluatedMsg{err: err}
		}
		return execEvaluatedMsg{output: output, finalLine: finalOutputLine(output)}
	}
}

func (m Model) readOptionalAgentOutputCommand() tea.Cmd {
	return func() tea.Msg {
		workspace := Workspace{TempDir: m.context.TempDir, Artifact: m.context.Artifact}
		output, err := workspace.readAgentOutput(false)
		return interactiveOutputMsg{output: output, err: err}
	}
}

func (m Model) prepareInteractiveCommand() tea.Cmd {
	return func() tea.Msg {
		workspace := Workspace{TempDir: m.context.TempDir, Artifact: m.context.Artifact}
		session, err := workspace.readSession()
		return interactivePreparedMsg{sessionID: session, err: err}
	}
}

func (m Model) completeStepCommand() tea.Cmd {
	return func() tea.Msg {
		workspace := Workspace{TempDir: m.context.TempDir, Artifact: m.context.Artifact}
		content, identical, err := workspace.compare(m.baseline)
		if err != nil {
			return stepCompletedMsg{err: err}
		}
		if identical {
			return stepCompletedMsg{}
		}
		if err := m.store.Save(m.context.ChangeID, m.context.Artifact, content); err != nil {
			return stepCompletedMsg{err: fmt.Errorf("save %s artifact: %w", m.context.Artifact, err)}
		}
		return stepCompletedMsg{}
	}
}

func (m Model) renderCommand(mode PreviewMode) tea.Cmd {
	if m.context.Artifact != ArtifactIdea && m.context.Artifact != ArtifactSpec && m.context.Artifact != ArtifactPR {
		return func() tea.Msg {
			return renderedMsg{mode: mode, err: fmt.Errorf("preview does not support artifact %q", m.context.Artifact)}
		}
	}
	workspace := Workspace{TempDir: m.context.TempDir, Artifact: m.context.Artifact}
	directory, err := workspace.Directory()
	if err != nil {
		return func() tea.Msg { return renderedMsg{mode: mode, err: err} }
	}
	input, inputErr := workspace.InputPath()
	if inputErr != nil {
		return func() tea.Msg { return renderedMsg{mode: mode, err: inputErr} }
	}
	output, outputErr := workspace.OutputPath()
	if outputErr != nil {
		return func() tea.Msg { return renderedMsg{mode: mode, err: outputErr} }
	}
	return func() tea.Msg {
		if _, err := os.ReadFile(input); err != nil {
			return renderedMsg{mode: mode, err: fmt.Errorf("read input.md before Preview: %w", err)}
		}
		if _, err := os.ReadFile(output); err != nil {
			return renderedMsg{mode: mode, err: fmt.Errorf("read output.md before Preview: %w", err)}
		}
		operationContext := context.Background()
		theme := m.screens[m.screen].Options.Theme
		if mode == PreviewArtifact {
			command := m.operations.Preview(operationContext, output, directory, theme, func(result RenderResult, err error) tea.Msg {
				return renderedMsg{mode: mode, result: result, err: err}
			})
			if command == nil {
				return renderedMsg{mode: mode, err: fmt.Errorf("preview operation returned no command")}
			}
			return command()
		}
		command := m.operations.Diff(operationContext, input, output, directory, theme, func(result RenderResult, err error) tea.Msg {
			return renderedMsg{mode: mode, result: result, err: err}
		})
		if command == nil {
			return renderedMsg{mode: mode, err: fmt.Errorf("diff operation returned no command")}
		}
		return command()
	}
}

func (m Model) activeTask() (TaskDefinition, error) {
	step, exists := m.steps[m.context.Step]
	if !exists || m.taskIndex < 0 || m.taskIndex >= len(step.Tasks) {
		return TaskDefinition{}, fmt.Errorf("current Flow task is unavailable")
	}
	return step.Tasks[m.taskIndex], nil
}

func (m Model) activeInteractiveTask() (TaskDefinition, error) {
	task, err := m.activeTask()
	if err == nil && task.Type == TaskInteractive {
		return task, nil
	}
	for _, step := range m.definition.Steps {
		for _, candidate := range step.Tasks {
			if candidate.Type == TaskInteractive && candidate.Screen == m.screen && candidate.Artifact == m.context.Artifact {
				return candidate, nil
			}
		}
	}
	return TaskDefinition{}, fmt.Errorf("interactive Screen %q has no configured task for artifact %q", m.screen, m.context.Artifact)
}

func (m Model) withTaskError(err error) Model {
	task, taskErr := m.activeTask()
	if taskErr != nil {
		return m.withError(err, "")
	}
	return m.withError(err, task.Error)
}

func (m Model) withError(err error, screen ScreenID) Model {
	if m.cancel != nil {
		m.cancel()
	}
	m.cancel = nil
	m.running = false
	m.err = err
	m.rendered = ""
	if screen != "" {
		m.screen = screen
	} else {
		m.screen = firstScreenOfType(m.definition, ScreenError)
	}
	return m
}

func (m Model) finish(screen ScreenID) Model {
	m.done = true
	m.terminalScreen = screen
	m.screen = ""
	m.running = false
	m.cancel = nil
	return m
}

func firstScreenOfType(definition Definition, screenType ScreenType) ScreenID {
	for _, screen := range definition.Screens {
		if screen.Type == screenType {
			return screen.ID
		}
	}
	return ""
}

func cloneDefinition(definition Definition) Definition {
	clone := Definition{ID: definition.ID}
	clone.Steps = make([]StepDefinition, len(definition.Steps))
	for index, step := range definition.Steps {
		clone.Steps[index] = StepDefinition{ID: step.ID, Tasks: append([]TaskDefinition(nil), step.Tasks...)}
	}
	clone.Screens = make([]ScreenDefinition, len(definition.Screens))
	for index, screen := range definition.Screens {
		clone.Screens[index] = screen
		clone.Screens[index].Commands = append([]CommandDefinition(nil), screen.Commands...)
	}
	return clone
}

func finalOutputLine(output string) string {
	normalized := strings.ReplaceAll(output, "\r\n", "\n")
	normalized = strings.TrimRight(normalized, "\n")
	if index := strings.LastIndex(normalized, "\n"); index >= 0 {
		return normalized[index+1:]
	}
	return normalized
}

func joinView(parts ...string) string {
	nonEmpty := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			nonEmpty = append(nonEmpty, part)
		}
	}
	return strings.Join(nonEmpty, "\n")
}
