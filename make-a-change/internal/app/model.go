package app

import (
	"strconv"

	"mch/internal/changes"
	"mch/internal/dto"
	"mch/internal/epics"
	"mch/internal/projects"
	"mch/internal/styles"
	httpclient "mch/pkg/client"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const defaultBackendURL = "http://localhost:8080"
const noProjectsToSelectError = "No projects to select from. Please create new project and select it on Main Screen."
const defaultInputPlaceholder = "Type / for commands"

type dropdownKind string

const (
	dropdownCommand dropdownKind = "command"
	dropdownList    dropdownKind = "list"
	dropdownSelect  dropdownKind = "select"
	dropdownConfirm dropdownKind = "confirm"
)

type selectorSource string

const (
	selectorProjects selectorSource = "projects"
	selectorPhases   selectorSource = "phases"
	selectorEpics    selectorSource = "epics"
	selectorTypes    selectorSource = "types"
)

type filterField string

const (
	filterPhase filterField = "phase"
	filterEpic  filterField = "epic"
	filterType  filterField = "type"
)

type changesFilters struct {
	phase dto.Option
	epic  dto.Option
	typ   dto.Option
}

type dropdownModel struct {
	kind        dropdownKind
	state       State
	previous    State
	onSelect    State
	source      selectorSource
	filterField filterField
	label       string
	options     []dto.Option
	filter      string
	highlighted int
	loading     bool
}

type selectorLoadedMsg struct {
	source  selectorSource
	options []dto.Option
	err     error
}

type projectListLoadedMsg struct {
	projects []dto.Project
	err      error
}

type projectSavedMsg struct {
	source  State
	project dto.Project
	err     error
}

type projectLoadedMsg struct {
	id      int
	project dto.Project
	err     error
}

type currentProjectLoadedMsg struct {
	id      int
	project dto.Project
	err     error
}

type editorFinishedMsg struct {
	source  State
	content string
	err     error
}

type startupProjectSelectionMsg struct{}

type appClient interface {
	projects.API
	changes.API
	epics.API
}

// Model is the root Bubble Tea model for the mch application shell.
type Model struct {
	input           textarea.Model
	state           State
	previousState   State
	width           int
	quitting        bool
	err             string
	status          string
	helpQuery       string
	promptCursorRow int
	promptCursorCol int
	pendingAltO     bool
	changesFilters  changesFilters
	currentProject  dto.Option
	projectList     projects.Model
	client          appClient
	appConfig       appConfig
	configPath      string
	dropdown        dropdownModel
}

// NewModel creates the default mch model using local config and HTTP backend access.
func NewModel() Model {
	configPath := resolveConfigPath(defaultConfigPath)
	cfg, err := loadAppConfig(configPath)
	m := newModelWithConfig(httpclient.NewHTTPClient(cfg.BackendURL), cfg, configPath)
	if err != nil {
		m.err = err.Error()
	}
	return m
}

// NewModelWithClient creates a model with an injected backend client for tests.
func NewModelWithClient(client appClient) Model {
	return newModelWithConfig(client, appConfig{BackendURL: defaultBackendURL}, "")
}

func newModelWithConfig(client appClient, cfg appConfig, configPath string) Model {
	input := textarea.New()
	input.Placeholder = defaultInputPlaceholder
	input.Prompt = "> "
	input.ShowLineNumbers = false
	input.EndOfBufferCharacter = ' '
	input.CharLimit = 240
	input.SetWidth(0)
	input.SetHeight(1)
	input.FocusedStyle.Base = styles.Default.InputBand
	input.FocusedStyle.Prompt = styles.Default.InputBand.Foreground(lipgloss.Color("183"))
	input.FocusedStyle.Text = styles.Default.InputBand.Foreground(lipgloss.Color("15"))
	input.FocusedStyle.CursorLine = styles.Default.InputBand.Foreground(lipgloss.Color("15"))
	input.FocusedStyle.Placeholder = styles.Default.InputBand.Foreground(lipgloss.Color("0"))
	input.FocusedStyle.EndOfBuffer = styles.Default.InputBand.Foreground(lipgloss.Color("240"))
	input.BlurredStyle = input.FocusedStyle
	input.Cursor.Style = styles.Default.InputBand.Foreground(lipgloss.Color("15"))
	input.Cursor.TextStyle = input.FocusedStyle.Text
	input.Cursor.SetMode(cursor.CursorStatic)
	input.Focus()

	currentProject := dto.Option{}
	if cfg.ProjectID > 0 {
		currentProject = dto.Option{
			ID: strconv.Itoa(cfg.ProjectID),
		}
	}

	return Model{
		input:          input,
		state:          MainState,
		width:          80,
		currentProject: currentProject,
		client:         client,
		appConfig:      cfg,
		configPath:     configPath,
		status:         "MainState",
	}
}

var _ tea.Model = Model{}
