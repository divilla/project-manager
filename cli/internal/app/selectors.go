package app

import (
	"fmt"
	"strconv"

	"mch/internal/dto"

	tea "github.com/charmbracelet/bubbletea"
)

func filterOptions(options []dto.Option) []dto.Option {
	filtered := make([]dto.Option, 0, len(options)+1)
	filtered = append(filtered, options...)
	filtered = append(filtered, dto.Option{ID: "/clear", Label: "/clear"})
	return filtered
}

func (m *Model) setChangesFilter(field filterField, option dto.Option) {
	switch field {
	case filterPhase:
		m.changesFilters.phase = option
	case filterEpic:
		m.changesFilters.epic = option
	case filterType:
		m.changesFilters.typ = option
	}
}

func (m *Model) clearChangesFilter(field filterField) {
	switch field {
	case filterPhase:
		m.changesFilters.phase = dto.Option{}
	case filterEpic:
		m.changesFilters.epic = dto.Option{}
	case filterType:
		m.changesFilters.typ = dto.Option{}
	}
}

func (m *Model) saveCurrentProject(project dto.Option) error {
	if m.configPath == "" {
		return nil
	}
	projectID, err := strconv.Atoi(project.ID)
	if err != nil {
		return fmt.Errorf("failed to save project_id: current project ID must be numeric")
	}
	m.appConfig.ProjectID = projectID
	if m.appConfig.BackendURL == "" {
		m.appConfig.BackendURL = defaultBackendURL
	}
	if err := saveAppConfig(m.configPath, m.appConfig); err != nil {
		return fmt.Errorf("failed to save project_id: %w", err)
	}
	return nil
}

func selectorSourceForState(state State) selectorSource {
	switch state {
	case SelectProjectDropDown:
		return selectorProjects
	case SelectPhaseDropDown:
		return selectorPhases
	case SelectEpicDropDown:
		return selectorEpics
	case SelectTypesDropDown:
		return selectorTypes
	default:
		return ""
	}
}

func selectorCommand(client appClient, source selectorSource, projectID string) tea.Cmd {
	return func() tea.Msg {
		var (
			options []dto.Option
			err     error
		)
		switch source {
		case selectorProjects:
			options, err = client.ListProjects()
		case selectorEpics:
			options, err = client.ListEpics(projectID)
		case selectorPhases:
			options, err = client.ListPhases()
		case selectorTypes:
			options, err = client.ListTypes()
		}
		return selectorLoadedMsg{source: source, options: options, err: err}
	}
}

func projectListCommand(client appClient) tea.Cmd {
	return func() tea.Msg {
		projects, err := client.ListProjectRows()
		return projectListLoadedMsg{projects: projects, err: err}
	}
}
