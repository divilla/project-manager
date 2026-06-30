package app

import (
	"mch/internal/changes"
	"mch/internal/dto"
	"mch/internal/epics"
	"mch/internal/help"
	"mch/internal/planning"
	"mch/internal/projects"
	"mch/internal/testcases"
)

var commandsByState = map[State][]string{
	MainState:            planning.MainCommands(),
	ChangesListState:     changes.ListCommands(),
	ChangeDetailsState:   changes.DetailCommands(),
	TestCaseDetailsState: testcases.DetailCommands(),
	ChangeCreateState:    {"/save", "/cancel"},
	ChangeUpdateState:    {"/save", "/cancel"},
	TestCaseCreateState:  testcases.EditCommands(),
	TestCaseUpdateState:  testcases.EditCommands(),
	EpicsListState:       epics.ListCommands(),
	EpicDetailsState:     epics.DetailCommands(),
	EpicCreateState:      {"/save", "/cancel"},
	EpicUpdateState:      {"/save", "/cancel"},
	ProjectsListState:    projects.ListCommands(),
	ProjectDetailsState:  projects.DetailCommands(),
	ProjectCreateState:   {"/editor", "/save", "/cancel"},
	ProjectUpdateState:   {"/editor", "/save", "/cancel"},
	MainHelpState:        help.Commands(),
	ChangesHelpState:     help.Commands(),
	EpicsHelpState:       help.Commands(),
	ProjectsHelpState:    help.Commands(),
}

func commandOptions(state State) []dto.Option {
	commands := commandsByState[state]
	options := make([]dto.Option, 0, len(commands))
	for _, command := range commands {
		options = append(options, dto.Option{ID: command, Label: command})
	}
	return options
}

func commandAllowed(state State, command string) bool {
	for _, allowed := range commandsByState[state] {
		if allowed == command {
			return true
		}
	}
	return false
}

func helpStateFor(state State) State {
	switch state {
	case MainState:
		return MainHelpState
	case ChangesListState, ChangeDetailsState, TestCaseDetailsState:
		return ChangesHelpState
	case EpicsListState, EpicDetailsState:
		return EpicsHelpState
	case ProjectsListState, ProjectDetailsState:
		return ProjectsHelpState
	default:
		return state
	}
}
