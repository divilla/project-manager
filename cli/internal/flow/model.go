package flow

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Context contains runtime-only Flow state.
type Context struct {
	Root            string
	FlowDir         string
	ChangeID        int
	ChangeRef       string
	Origin          ScreenID
	Step            StepID
	Task            TaskID
	TaskIndex       int
	Artifact        Artifact
	SessionID       string
	ExecutionResult string
	FromStep        StepID
	EditorCaller    ScreenID
	WorkspaceScope  WorkspaceScope
}

// Composition supplies a definition and its runtime boundaries.
type Composition struct {
	Definition      Definition
	Context         Context
	TerminalScreens []ScreenID
	Store           ArtifactStore
	Options         DocumentOptions
	Operations      Operations
}

// CommandMsg selects a command on the active Flow Screen.
type CommandMsg struct{ ID CommandID }

// PreviewMode identifies artifact or diff rendering.
type PreviewMode string

// Supported Preview rendering modes.
const (
	PreviewArtifact PreviewMode = "preview"
	PreviewDiff     PreviewMode = "diff"
)

// Model is the reusable Bubble Tea Flow runtime.
type Model struct {
	definition Definition
	context    Context
	store      ArtifactStore
	options    DocumentOptions
	operations Operations
	steps      map[StepID]StepDefinition
	screens    map[ScreenID]ScreenDefinition
	terminals  map[ScreenID]struct{}

	screen         ScreenID
	previewMode    PreviewMode
	rendered       string
	baseline       []byte
	err            error
	validationErr  error
	chatLoading    bool
	running        bool
	done           bool
	terminalScreen ScreenID
	cancel         context.CancelFunc
	callerRendered string
	callerMode     PreviewMode
	callerOutput   string
	callerStep     StepID
	callerTask     TaskID
	callerIndex    int
	callerArtifact Artifact
	callerBaseline []byte
	callerScope    WorkspaceScope

	operationSequence uint64
	pendingOperation  uint64
}

type stepLoadedMsg struct {
	operation uint64
	stepID    StepID
	artifact  Artifact
	scope     WorkspaceScope
	baseline  []byte
	err       error
}

type chatLoadedMsg struct {
	operation uint64
	baseline  []byte
	output    string
	err       error
}

type editorFinishedMsg struct {
	operation uint64
	err       error
}
type editorPreparedMsg struct {
	operation uint64
	err       error
}
type execFinishedMsg struct {
	operation uint64
	err       error
}
type execEvaluatedMsg struct {
	operation uint64
	output    string
	finalLine string
	err       error
}
type chatPreparedMsg struct {
	operation uint64
	sessionID string
	err       error
}
type chatFinishedMsg struct {
	operation uint64
	err       error
}
type taskCompletedMsg struct {
	operation uint64
	empty     bool
	err       error
}
type renderedMsg struct {
	operation uint64
	mode      PreviewMode
	result    RenderResult
	err       error
}

// Compose validates and prepares an unstarted Flow model.
func Compose(composition Composition) Model {
	definition := cloneDefinition(composition.Definition)
	m := Model{
		definition:  definition,
		context:     composition.Context,
		store:       composition.Store,
		options:     composition.Options,
		operations:  composition.Operations,
		steps:       make(map[StepID]StepDefinition, len(definition.Steps)),
		screens:     make(map[ScreenID]ScreenDefinition, len(definition.Screens)),
		terminals:   make(map[ScreenID]struct{}, len(composition.TerminalScreens)),
		previewMode: PreviewArtifact,
	}
	for _, step := range definition.Steps {
		m.steps[step.ID] = step
	}
	for _, screen := range definition.Screens {
		m.screens[screen.ID] = screen
	}
	for _, terminal := range composition.TerminalScreens {
		m.terminals[terminal] = struct{}{}
	}
	if err := ValidateDefinition(definition, composition.TerminalScreens); err != nil {
		return m.withError(fmt.Errorf("invalid Flow definition: %w", err), "")
	}
	if m.store == nil {
		return m.withError(fmt.Errorf("artifact store is required"), "")
	}
	if m.operations == nil {
		return m.withError(fmt.Errorf("external operations boundary is required"), "")
	}
	if _, ok := m.terminals[m.context.Origin]; !ok {
		return m.withError(fmt.Errorf("flow context originating Screen %q is not an allowed terminal Screen", m.context.Origin), "")
	}
	if step, ok := m.steps[m.context.Step]; ok && step.Mode == ModeEditor {
		if err := m.validateEditorCaller(m.context.EditorCaller); err != nil {
			return m.withError(err, step.Tasks[0].Error)
		}
	}
	m.beginOperation()
	return m
}

// Init loads and starts the configured Step.
func (m Model) Init() tea.Cmd {
	if m.err != nil || m.done {
		return nil
	}
	return m.startStepCommand(m.context.Step, m.pendingOperation)
}

// Update applies operation results and navigation commands.
func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case stepLoadedMsg:
		if !m.completeOperation(message.operation) {
			return m, nil
		}
		return m.updateStepLoaded(message)
	case chatLoadedMsg:
		if !m.completeOperation(message.operation) {
			return m, nil
		}
		return m.updateChatLoaded(message)
	case editorFinishedMsg:
		if !m.completeOperation(message.operation) {
			return m, nil
		}
		return m.updateEditorFinished(message)
	case editorPreparedMsg:
		if !m.completeOperation(message.operation) {
			return m, nil
		}
		if message.err != nil {
			return m.withTaskError(message.err), nil
		}
		operation := m.beginOperation()
		return m, m.launchEditorCommand(operation)
	case execFinishedMsg:
		if !m.completeOperation(message.operation) {
			return m, nil
		}
		return m.updateExecFinished(message)
	case execEvaluatedMsg:
		if !m.completeOperation(message.operation) {
			return m, nil
		}
		return m.updateExecEvaluated(message)
	case chatPreparedMsg:
		if !m.completeOperation(message.operation) {
			return m, nil
		}
		return m.updateChatPrepared(message)
	case chatFinishedMsg:
		if !m.completeOperation(message.operation) {
			return m, nil
		}
		return m.updateChatFinished(message)
	case taskCompletedMsg:
		if !m.completeOperation(message.operation) {
			return m, nil
		}
		return m.updateTaskCompleted(message)
	case renderedMsg:
		if !m.completeOperation(message.operation) {
			return m, nil
		}
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

// View renders the active generic Screen.
func (m Model) View() string {
	if m.done {
		return ""
	}
	screen := m.screens[m.screen]
	title := screen.Title
	if title == "" {
		title = string(screen.Type)
	}
	if m.err != nil {
		return joinView(title, m.err.Error(), "/return")
	}
	if m.validationErr != nil {
		return joinView(title, m.validationErr.Error(), "/fix  /cancel")
	}
	switch screen.Type {
	case ScreenExec:
		return joinView(title, "/stop")
	case ScreenChat:
		labels := make([]string, 0, len(m.Commands()))
		for _, command := range m.Commands() {
			labels = append(labels, string(command))
		}
		return joinView(title, m.context.ExecutionResult, strings.Join(labels, "  "))
	case ScreenPreview:
		labels := make([]string, 0, len(m.Commands()))
		for _, command := range m.Commands() {
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

// Screen returns the active runtime Screen.
func (m Model) Screen() ScreenID { return m.screen }

// FlowContext returns a copy of current runtime state.
func (m Model) FlowContext() Context { return m.context }

// Error returns the current runtime error.
func (m Model) Error() error { return m.err }

// Rendered returns the latest Preview or Diff output.
func (m Model) Rendered() string { return m.rendered }

// Mode returns the active Preview rendering mode.
func (m Model) Mode() PreviewMode { return m.previewMode }

// Done reports whether the Flow reached a terminal Screen.
func (m Model) Done() bool { return m.done }

// TerminalScreen returns the selected terminal Screen.
func (m Model) TerminalScreen() ScreenID { return m.terminalScreen }

// Definition returns a copy of the active definition.
func (m Model) Definition() Definition { return cloneDefinition(m.definition) }

// ValidationError returns the recoverable Editor content error.
func (m Model) ValidationError() error { return m.validationErr }

// ActiveStep returns the current Step definition.
func (m Model) ActiveStep() StepDefinition { return m.steps[m.context.Step] }

// ActiveScreen returns the current Screen definition.
func (m Model) ActiveScreen() ScreenDefinition { return m.screens[m.screen] }

// Commands returns commands available on the active Screen.
func (m Model) Commands() []CommandID {
	if m.done {
		return nil
	}
	if m.pendingOperation != 0 {
		if m.screens[m.screen].Type == ScreenExec && m.running {
			return []CommandID{"/stop"}
		}
		return nil
	}
	if m.err != nil {
		return []CommandID{"/return"}
	}
	if m.validationErr != nil {
		return []CommandID{"/fix", "/cancel"}
	}
	screen := m.screens[m.screen]
	switch screen.Type {
	case ScreenExec:
		if m.running {
			return []CommandID{"/stop"}
		}
	case ScreenChat:
		if m.chatLoading {
			return nil
		}
		return []CommandID{"/chat", "/edit", "/cancel"}
	case ScreenPreview:
		commands := make([]CommandID, 0, len(screen.Commands)+1)
		for _, command := range screen.Commands {
			commands = append(commands, command.ID)
			if command.ID == "/continue" && m.previewHasChat() && !previewDeclares(screen, "/chat") {
				commands = append(commands, "/chat")
			}
		}
		return commands
	}
	return nil
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
	m.context.WorkspaceScope = message.scope
	m.context.TaskIndex = 0
	m.context.ExecutionResult = ""
	m.context.FromStep = ""
	m.baseline = append([]byte(nil), message.baseline...)
	m.previewMode = PreviewArtifact
	m.rendered = ""
	m.validationErr = nil
	return m.enterTask(step.Tasks[0])
}

func (m Model) updateChatLoaded(message chatLoadedMsg) (tea.Model, tea.Cmd) {
	if message.err != nil {
		return m.withTaskError(message.err), nil
	}
	m.chatLoading = false
	m.baseline = append([]byte(nil), message.baseline...)
	m.context.ExecutionResult = message.output
	return m, nil
}

func (m Model) updateEditorFinished(message editorFinishedMsg) (tea.Model, tea.Cmd) {
	if m.done || m.screens[m.screen].Type != ScreenEditor {
		return m, nil
	}
	if message.err != nil {
		return m.withTaskError(fmt.Errorf("editor failed: %w", message.err)), nil
	}
	m.validationErr = nil
	operation := m.beginOperation()
	return m, m.completeTaskCommand(SaveByUser, operation)
}

func (m Model) updateExecFinished(message execFinishedMsg) (tea.Model, tea.Cmd) {
	if m.done || m.screens[m.screen].Type != ScreenExec {
		return m, nil
	}
	m.stopRunning()
	if message.err != nil {
		return m.withTaskError(fmt.Errorf("execution failed: %w", message.err)), nil
	}
	operation := m.beginOperation()
	return m, m.completeTaskCommand(SaveByAgent, operation)
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
		return m.advanceTo(m.context.TaskIndex + 2)
	}
	return m.advanceTo(m.context.TaskIndex + 1)
}

func (m Model) updateChatPrepared(message chatPreparedMsg) (tea.Model, tea.Cmd) {
	if message.err != nil {
		return m.withTaskError(message.err), nil
	}
	m.context.SessionID = message.sessionID
	task, err := m.activeChatTask()
	if err != nil {
		return m.withTaskError(err), nil
	}
	workspace := m.workspace()
	directory, err := workspace.Directory()
	if err != nil {
		return m.withTaskError(err), nil
	}
	operationContext, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	m.running = true
	request := ChatRequest{Script: task.Script, FlowDir: m.context.FlowDir, Workspace: directory, SessionID: message.sessionID, ChangeID: m.context.ChangeID, ChangeRef: m.context.ChangeRef, Artifact: task.Artifact}
	operation := m.beginOperation()
	return m, m.operations.Chat(operationContext, request, func(err error) tea.Msg {
		return chatFinishedMsg{operation: operation, err: err}
	})
}

func (m Model) updateChatFinished(message chatFinishedMsg) (tea.Model, tea.Cmd) {
	if m.done || m.screens[m.screen].Type != ScreenChat {
		return m, nil
	}
	m.stopRunning()
	if message.err != nil {
		return m.withTaskError(fmt.Errorf("chat session failed: %w", message.err)), nil
	}
	operation := m.beginOperation()
	return m, m.completeTaskCommand(SaveByAgent, operation)
}

func (m Model) updateTaskCompleted(message taskCompletedMsg) (tea.Model, tea.Cmd) {
	if message.err != nil {
		var validation ValidationError
		if m.screens[m.screen].Type == ScreenEditor && errors.As(message.err, &validation) {
			m.validationErr = validation
			return m, nil
		}
		return m.withTaskError(message.err), nil
	}
	task, err := m.activeTask()
	if err != nil {
		return m.withTaskError(err), nil
	}
	if m.screens[m.screen].Type == ScreenEditor {
		if message.empty {
			if task.Type == TaskEditor {
				return m.finish(task.Cancel.Screen), nil
			}
			return m.finish(m.context.Origin), nil
		}
		if m.outputEqualsBaseline() {
			return m.finishOrReturnCaller(), nil
		}
		if task.Type == TaskEditor {
			m.context.WorkspaceScope = WorkspaceArtifact
		}
		return m.enterPreview(task.Preview)
	}
	switch task.Type {
	case TaskExec:
		operation := m.beginOperation()
		return m, m.evaluateExecCommand(operation)
	case TaskChat:
		return m.advanceTo(m.context.TaskIndex + 1)
	default:
		return m.withTaskError(fmt.Errorf("task %q cannot complete from Screen %q", task.ID, m.screen)), nil
	}
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
	if m.pendingOperation != 0 {
		if m.screens[m.screen].Type == ScreenExec && command == "/stop" && m.running {
			m.stopRunning()
			m.pendingOperation = 0
			return m.finish(m.context.Origin), nil
		}
		return m, nil
	}
	if m.err != nil {
		if command == "/return" {
			return m.finish(m.context.Origin), nil
		}
		return m, nil
	}
	if m.validationErr != nil {
		switch command {
		case "/fix":
			m.validationErr = nil
			operation := m.beginOperation()
			return m, m.prepareEditorCommand(operation)
		case "/cancel":
			task, err := m.activeTask()
			if err != nil {
				return m.withTaskError(err), nil
			}
			if task.Type == TaskEditor {
				return m.finish(task.Cancel.Screen), nil
			}
			return m.finish(m.context.Origin), nil
		}
		return m, nil
	}
	screen := m.screens[m.screen]
	switch screen.Type {
	case ScreenExec:
		if command == "/stop" && m.running {
			m.stopRunning()
			return m.finish(m.context.Origin), nil
		}
	case ScreenChat:
		if m.chatLoading {
			return m, nil
		}
		switch command {
		case "/cancel":
			m.stopRunning()
			return m.finish(m.context.Origin), nil
		case "/chat":
			operation := m.beginOperation()
			return m, m.prepareChatCommand(operation)
		case "/edit":
			task, err := m.activeChatTask()
			if err != nil {
				return m.withTaskError(err), nil
			}
			m.rememberEditorCaller()
			if task.Editor.Kind == DestinationStep {
				m.context.Step = task.Editor.Step
				operation := m.beginOperation()
				return m, m.startStepCommand(task.Editor.Step, operation)
			}
			m.screen = task.Editor.Screen
			operation := m.beginOperation()
			return m, m.prepareEditorCommand(operation)
		}
	case ScreenPreview:
		if command == "/chat" && m.previewHasChat() {
			step := m.steps[screen.FromStep]
			for index, task := range step.Tasks {
				if task.Type == TaskChat {
					m.context.Step = step.ID
					m.context.TaskIndex = index
					return m.enterTask(task)
				}
			}
		}
		for _, defined := range screen.Commands {
			if defined.ID != command || command == "/chat" {
				continue
			}
			if defined.Destination.Kind == DestinationStep {
				if command == "/edit" {
					m.rememberEditorCaller()
				} else {
					m.context.EditorCaller = ""
				}
				m.context.Step = defined.Destination.Step
				operation := m.beginOperation()
				return m, m.startStepCommand(defined.Destination.Step, operation)
			}
			return m.finish(defined.Destination.Screen), nil
		}
	}
	return m, nil
}

func (m Model) togglePreview() (tea.Model, tea.Cmd) {
	if m.err != nil || m.done || m.pendingOperation != 0 || m.screens[m.screen].Type != ScreenPreview {
		return m, nil
	}
	mode := PreviewArtifact
	if m.previewMode == PreviewArtifact {
		mode = PreviewDiff
	}
	operation := m.beginOperation()
	return m, m.renderCommand(mode, operation)
}

func (m Model) enterTask(task TaskDefinition) (tea.Model, tea.Cmd) {
	m.context.Task = task.ID
	m.context.Artifact = task.Artifact
	m.screen = task.Screen
	m.err = nil
	m.validationErr = nil
	switch task.Type {
	case TaskEditor:
		if err := m.validateEditorCaller(m.context.EditorCaller); err != nil {
			return m.withTaskError(err), nil
		}
		operation := m.beginOperation()
		return m, m.prepareEditorCommand(operation)
	case TaskExec:
		return m.startExec(task)
	case TaskChat:
		m.chatLoading = true
		operation := m.beginOperation()
		return m, m.refreshChatCommand(operation)
	default:
		return m.withTaskError(fmt.Errorf("unsupported task type %q", task.Type)), nil
	}
}

func (m Model) startExec(task TaskDefinition) (tea.Model, tea.Cmd) {
	directory, err := m.workspace().Directory()
	if err != nil {
		return m.withTaskError(err), nil
	}
	operationContext, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	m.running = true
	request := ExecRequest{Script: task.Script, Prompt: task.Prompt, FlowDir: m.context.FlowDir, Workspace: directory, ChangeID: m.context.ChangeID, ChangeRef: m.context.ChangeRef, Artifact: task.Artifact}
	operation := m.beginOperation()
	return m, m.operations.Exec(operationContext, request, func(err error) tea.Msg {
		return execFinishedMsg{operation: operation, err: err}
	})
}

func (m Model) advanceTo(index int) (tea.Model, tea.Cmd) {
	step := m.steps[m.context.Step]
	if index >= len(step.Tasks) {
		task, err := m.activeTask()
		if err != nil {
			return m.withTaskError(err), nil
		}
		return m.enterPreview(task.Preview)
	}
	m.context.TaskIndex = index
	return m.enterTask(step.Tasks[index])
}

func (m Model) enterPreview(preview ScreenID) (tea.Model, tea.Cmd) {
	m.context.FromStep = m.context.Step
	m.screen = preview
	m.previewMode = PreviewArtifact
	m.rendered = ""
	operation := m.beginOperation()
	return m, m.renderCommand(PreviewArtifact, operation)
}

func (m Model) startStepCommand(stepID StepID, operation uint64) tea.Cmd {
	return func() tea.Msg {
		step, exists := m.steps[stepID]
		if !exists {
			return stepLoadedMsg{operation: operation, stepID: stepID, err: fmt.Errorf("flow context references unknown Step %q", stepID)}
		}
		if m.context.ChangeID <= 0 {
			return stepLoadedMsg{operation: operation, stepID: stepID, err: fmt.Errorf("flow context Change ID must be a valid positive number")}
		}
		firstTask := step.Tasks[0]
		artifact := firstTask.Artifact
		scope := WorkspaceArtifact
		if firstTask.Type == TaskEditor {
			scope = WorkspaceEditor
		}
		workspace := Workspace{Root: m.context.Root, ChangeRef: m.context.ChangeRef, Artifact: artifact, Scope: scope}
		if _, err := workspace.Directory(); err != nil {
			return stepLoadedMsg{operation: operation, stepID: stepID, artifact: artifact, scope: scope, err: err}
		}
		content, err := m.store.Load(m.context.ChangeID, artifact)
		if err != nil {
			return stepLoadedMsg{operation: operation, stepID: stepID, artifact: artifact, scope: scope, err: fmt.Errorf("load %s artifact: %w", artifact, err)}
		}
		if err := workspace.replaceBaseline(content); err != nil {
			return stepLoadedMsg{operation: operation, stepID: stepID, artifact: artifact, scope: scope, err: err}
		}
		return stepLoadedMsg{operation: operation, stepID: stepID, artifact: artifact, scope: scope, baseline: append([]byte(nil), content...)}
	}
}

func (m Model) refreshChatCommand(operation uint64) tea.Cmd {
	return func() tea.Msg {
		content, err := m.store.Load(m.context.ChangeID, m.context.Artifact)
		if err != nil {
			return chatLoadedMsg{operation: operation, err: fmt.Errorf("load %s artifact for Chat: %w", m.context.Artifact, err)}
		}
		workspace := m.workspace()
		if err := workspace.replaceBaseline(content); err != nil {
			return chatLoadedMsg{operation: operation, err: err}
		}
		output, err := workspace.readAgentOutput(false)
		return chatLoadedMsg{operation: operation, baseline: append([]byte(nil), content...), output: output, err: err}
	}
}

func (m Model) completeTaskCommand(provenance SaveProvenance, operation uint64) tea.Cmd {
	return func() tea.Msg {
		workspace := m.workspace()
		empty, err := workspace.outputIsEmpty()
		if err != nil {
			return taskCompletedMsg{operation: operation, err: err}
		}
		if empty && m.screens[m.screen].Type == ScreenEditor {
			return taskCompletedMsg{operation: operation, empty: true}
		}
		content, identical, err := workspace.canonicalize(m.baseline, m.options)
		if err != nil {
			return taskCompletedMsg{operation: operation, err: err}
		}
		if identical {
			return taskCompletedMsg{operation: operation}
		}
		if err := m.store.Save(m.context.ChangeID, m.context.Artifact, content, provenance); err != nil {
			return taskCompletedMsg{operation: operation, err: fmt.Errorf("save %s artifact: %w", m.context.Artifact, err)}
		}
		task, err := m.activeTask()
		if err != nil {
			return taskCompletedMsg{operation: operation, err: err}
		}
		if task.Type == TaskEditor {
			if err := workspace.publishEditorPreview(); err != nil {
				return taskCompletedMsg{operation: operation, err: err}
			}
		}
		return taskCompletedMsg{operation: operation}
	}
}

func (m Model) evaluateExecCommand(operation uint64) tea.Cmd {
	return func() tea.Msg {
		output, err := m.workspace().readAgentOutput(true)
		if err != nil {
			return execEvaluatedMsg{operation: operation, err: err}
		}
		return execEvaluatedMsg{operation: operation, output: output, finalLine: finalOutputLine(output)}
	}
}

func (m Model) prepareEditorCommand(operation uint64) tea.Cmd {
	return func() tea.Msg {
		output, err := m.workspace().OutputPath()
		if err != nil {
			return editorPreparedMsg{operation: operation, err: err}
		}
		if _, err := os.ReadFile(output); err != nil {
			return editorPreparedMsg{operation: operation, err: fmt.Errorf("read output.md before Editor: %w", err)}
		}
		return editorPreparedMsg{operation: operation}
	}
}

func (m Model) launchEditorCommand(operation uint64) tea.Cmd {
	workspace := m.workspace()
	output, err := workspace.OutputPath()
	if err != nil {
		return func() tea.Msg { return editorFinishedMsg{operation: operation, err: err} }
	}
	directory, err := workspace.Directory()
	if err != nil {
		return func() tea.Msg { return editorFinishedMsg{operation: operation, err: err} }
	}
	return m.operations.Editor(output, directory, func(err error) tea.Msg {
		return editorFinishedMsg{operation: operation, err: err}
	})
}

func (m Model) prepareChatCommand(operation uint64) tea.Cmd {
	return func() tea.Msg {
		session, err := m.workspace().readSession()
		return chatPreparedMsg{operation: operation, sessionID: session, err: err}
	}
}

func (m Model) renderCommand(mode PreviewMode, operation uint64) tea.Cmd {
	if m.context.Artifact != ArtifactIdea && m.context.Artifact != ArtifactSpec && m.context.Artifact != ArtifactPR {
		return func() tea.Msg {
			return renderedMsg{operation: operation, mode: mode, err: fmt.Errorf("preview does not support artifact %q", m.context.Artifact)}
		}
	}
	workspace := m.workspace()
	directory, err := workspace.Directory()
	if err != nil {
		return func() tea.Msg { return renderedMsg{operation: operation, mode: mode, err: err} }
	}
	input, err := workspace.InputPath()
	if err != nil {
		return func() tea.Msg { return renderedMsg{operation: operation, mode: mode, err: err} }
	}
	output, err := workspace.OutputPath()
	if err != nil {
		return func() tea.Msg { return renderedMsg{operation: operation, mode: mode, err: err} }
	}
	return func() tea.Msg {
		if _, err := os.ReadFile(input); err != nil {
			return renderedMsg{operation: operation, mode: mode, err: fmt.Errorf("read input.md before Preview: %w", err)}
		}
		if _, err := os.ReadFile(output); err != nil {
			return renderedMsg{operation: operation, mode: mode, err: fmt.Errorf("read output.md before Preview: %w", err)}
		}
		theme := m.screens[m.screen].Options.Theme
		var command tea.Cmd
		if mode == PreviewArtifact {
			command = m.operations.Preview(context.Background(), output, directory, theme, func(result RenderResult, err error) tea.Msg {
				return renderedMsg{operation: operation, mode: mode, result: result, err: err}
			})
		} else {
			command = m.operations.Diff(context.Background(), input, output, directory, theme, func(result RenderResult, err error) tea.Msg {
				return renderedMsg{operation: operation, mode: mode, result: result, err: err}
			})
		}
		if command == nil {
			return renderedMsg{operation: operation, mode: mode, err: fmt.Errorf("%s operation returned no command", mode)}
		}
		return command()
	}
}

func (m Model) activeTask() (TaskDefinition, error) {
	step, exists := m.steps[m.context.Step]
	if !exists || m.context.TaskIndex < 0 || m.context.TaskIndex >= len(step.Tasks) {
		return TaskDefinition{}, fmt.Errorf("current Flow task is unavailable")
	}
	return step.Tasks[m.context.TaskIndex], nil
}

func (m Model) activeChatTask() (TaskDefinition, error) {
	task, err := m.activeTask()
	if err == nil && task.Type == TaskChat {
		return task, nil
	}
	return TaskDefinition{}, fmt.Errorf("Chat Screen %q has no active Chat task", m.screen)
}

func (m Model) workspace() Workspace {
	return Workspace{Root: m.context.Root, ChangeRef: m.context.ChangeRef, Artifact: m.context.Artifact, Scope: m.context.WorkspaceScope}
}

func (m Model) outputEqualsBaseline() bool {
	content, err := os.ReadFile(mustPath(m.workspace().OutputPath()))
	return err == nil && string(content) == string(m.baseline)
}

func (m Model) finishOrReturnCaller() Model {
	if _, terminal := m.terminals[m.context.EditorCaller]; terminal {
		return m.finish(m.context.EditorCaller)
	}
	m.screen = m.context.EditorCaller
	m.rendered = m.callerRendered
	m.previewMode = m.callerMode
	m.context.ExecutionResult = m.callerOutput
	m.context.Step = m.callerStep
	m.context.Task = m.callerTask
	m.context.TaskIndex = m.callerIndex
	m.context.Artifact = m.callerArtifact
	m.context.WorkspaceScope = m.callerScope
	m.baseline = append([]byte(nil), m.callerBaseline...)
	if m.screens[m.screen].Type == ScreenPreview {
		m.context.FromStep = m.screens[m.screen].FromStep
	}
	return m
}

func (m *Model) rememberEditorCaller() {
	m.context.EditorCaller = m.screen
	m.callerRendered = m.rendered
	m.callerMode = m.previewMode
	m.callerOutput = m.context.ExecutionResult
	m.callerStep = m.context.Step
	m.callerTask = m.context.Task
	m.callerIndex = m.context.TaskIndex
	m.callerArtifact = m.context.Artifact
	m.callerBaseline = append([]byte(nil), m.baseline...)
	m.callerScope = m.context.WorkspaceScope
}

func (m Model) validateEditorCaller(caller ScreenID) error {
	if _, ok := m.terminals[caller]; ok {
		return nil
	}
	screen, ok := m.screens[caller]
	if !ok || (screen.Type != ScreenPreview && screen.Type != ScreenChat) {
		return fmt.Errorf("Editor caller %q must be an allowed terminal, Preview, or Chat Screen", caller)
	}
	return nil
}

func (m Model) previewHasChat() bool {
	screen := m.screens[m.screen]
	step, ok := m.steps[screen.FromStep]
	if !ok || (step.Mode != ModeExec && step.Mode != ModeChat) {
		return false
	}
	_, ok = chatTask(step)
	return ok
}

func previewDeclares(screen ScreenDefinition, command CommandID) bool {
	for _, candidate := range screen.Commands {
		if candidate.ID == command {
			return true
		}
	}
	return false
}

func (m *Model) beginOperation() uint64 {
	m.operationSequence++
	m.pendingOperation = m.operationSequence
	return m.pendingOperation
}

func (m *Model) completeOperation(operation uint64) bool {
	if operation == 0 || operation != m.pendingOperation {
		return false
	}
	m.pendingOperation = 0
	return true
}

func (m Model) withTaskError(err error) Model {
	task, taskErr := m.activeTask()
	if taskErr != nil {
		return m.withError(err, "")
	}
	return m.withError(err, task.Error)
}

func (m Model) withError(err error, screen ScreenID) Model {
	m.stopRunning()
	m.pendingOperation = 0
	m.err = err
	m.validationErr = nil
	m.rendered = ""
	if screen != "" {
		m.screen = screen
	} else {
		m.screen = firstScreenOfType(m.definition, ScreenError)
	}
	return m
}

func (m *Model) stopRunning() {
	if m.cancel != nil {
		m.cancel()
	}
	m.cancel = nil
	m.running = false
}

func (m Model) finish(screen ScreenID) Model {
	m.stopRunning()
	m.pendingOperation = 0
	m.done = true
	m.terminalScreen = screen
	m.screen = ""
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
	clone := Definition{ID: definition.ID, Steps: make([]StepDefinition, len(definition.Steps)), Screens: make([]ScreenDefinition, len(definition.Screens))}
	for index, step := range definition.Steps {
		clone.Steps[index] = StepDefinition{ID: step.ID, Mode: step.Mode, Tasks: append([]TaskDefinition(nil), step.Tasks...)}
	}
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

func mustPath(path string, err error) string {
	if err != nil {
		return ""
	}
	return path
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
