package app

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"mch/internal/changes"
	"mch/internal/dto"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) saveChangeCreate() (tea.Model, tea.Cmd) {
	return m.saveChangeCreateValue(m.input.Value())
}

func (m Model) saveChangeCreateValue(body string) (tea.Model, tea.Cmd) {
	projectID, err := currentProjectNumericID(m.currentProject.ID)
	if err != nil {
		m.err = err.Error()
		m.status = "validation failed"
		return m, nil
	}
	m.status = "saving"
	return m, changeCreateCommand(m.client, projectID, m.currentProject.ID, body)
}

func (m Model) saveChangeUpdate() (tea.Model, tea.Cmd) {
	return m.saveChangeUpdateValue(m.input.Value())
}

func (m Model) saveChangeUpdateValue(body string) (tea.Model, tea.Cmd) {
	id, err := changeNumericID(m.changeList.Detail)
	if err != nil {
		m.err = err.Error()
		m.status = "validation failed"
		return m, nil
	}
	m.status = "saving"
	return m, changeUpdateCommand(m.client, id, m.currentProject.ID, m.changeList.Detail, body)
}

func changeCreateCommand(client appClient, projectID int, projectIDValue string, body string) tea.Cmd {
	return func() tea.Msg {
		if _, err := changes.ParseBodyStructure(body); err != nil {
			return changeSavedMsg{source: ChangeCreateState, err: err}
		}
		types, epics, err := changeReferenceData(client, projectIDValue, body)
		if err != nil {
			return changeSavedMsg{source: ChangeCreateState, err: err}
		}
		parsed, err := changes.ParseBody(body, types, epics)
		if err != nil {
			return changeSavedMsg{source: ChangeCreateState, err: err}
		}
		created, err := client.CreateChange(dto.ChangeCreateInput{
			ProjectID:   projectID,
			Title:       parsed.Title,
			Body:        parsed.Body,
			ChangeTypes: parsed.ChangeTypes,
			EpicID:      parsed.EpicID,
		})
		if err != nil {
			return changeSavedMsg{source: ChangeCreateState, err: err}
		}
		id, err := changeNumericID(created)
		if err != nil {
			return changeSavedMsg{source: ChangeCreateState, err: err}
		}
		change, err := client.GetChange(id)
		if err != nil {
			return changeSavedMsg{source: ChangeCreateState, change: created, reloadErr: err}
		}
		return changeSavedMsg{source: ChangeCreateState, change: change}
	}
}

func changeUpdateCommand(client appClient, id int, projectID string, original dto.Change, body string) tea.Cmd {
	return func() tea.Msg {
		if _, err := changes.ParseBodyStructure(body); err != nil {
			return changeSavedMsg{source: ChangeUpdateState, err: err}
		}
		types, epics, err := changeReferenceData(client, projectID, body)
		if err != nil {
			return changeSavedMsg{source: ChangeUpdateState, err: err}
		}
		parsed, err := changes.ParseBody(body, types, epics)
		if err != nil {
			return changeSavedMsg{source: ChangeUpdateState, err: err}
		}
		if parsed.Title != original.Title {
			if _, err := client.UpdateChangeTitle(id, parsed.Title); err != nil {
				return changeSavedMsg{source: ChangeUpdateState, err: err}
			}
		}
		if parsed.Body != original.Body {
			if _, err := client.UpdateChangeBody(id, parsed.Body); err != nil {
				return changeSavedMsg{source: ChangeUpdateState, err: err}
			}
		}
		if !changes.SameTypes(parsed.ChangeTypes, original.ChangeTypes) {
			if _, err := client.UpdateChangeTypes(id, parsed.ChangeTypes); err != nil {
				return changeSavedMsg{source: ChangeUpdateState, err: err}
			}
		}
		if !sameEpicID(parsed.EpicID, original.EpicID) {
			if _, err := client.UpdateChangeEpic(id, parsed.EpicID); err != nil {
				return changeSavedMsg{source: ChangeUpdateState, err: err}
			}
		}
		change, err := client.GetChange(id)
		return changeSavedMsg{source: ChangeUpdateState, change: change, err: err}
	}
}

func changeGetCommand(client appClient, id int) tea.Cmd {
	return func() tea.Msg {
		change, err := client.GetChange(id)
		return changeLoadedMsg{id: id, change: change, err: err}
	}
}

func changeDeleteCommand(client appClient, change dto.Change, target State) tea.Cmd {
	return func() tea.Msg {
		id, err := changeNumericID(change)
		if err != nil {
			return changeDeletedMsg{target: target, err: err}
		}
		return changeDeletedMsg{target: target, err: client.DeleteChange(id)}
	}
}

func changeDetailFieldUpdateCommand(client appClient, change dto.Change, field detailEditField, selected dto.Option) tea.Cmd {
	return func() tea.Msg {
		id, err := changeNumericID(change)
		if err != nil {
			return changeSavedMsg{source: ChangeDetailsState, err: err}
		}
		switch field {
		case detailEditPhase:
			if _, err := client.UpdateChangePhase(id, selected.ID); err != nil {
				return changeSavedMsg{source: ChangeDetailsState, err: err}
			}
		case detailEditTypes:
			changeTypes := toggleChangeType(change.ChangeTypes, selected)
			if _, err := client.UpdateChangeTypes(id, changeTypes); err != nil {
				return changeSavedMsg{source: ChangeDetailsState, err: err}
			}
		case detailEditEpic:
			epicID, err := selectedEpicID(selected)
			if err != nil {
				return changeSavedMsg{source: ChangeDetailsState, err: err}
			}
			if _, err := client.UpdateChangeEpic(id, epicID); err != nil {
				return changeSavedMsg{source: ChangeDetailsState, err: err}
			}
		default:
			return changeSavedMsg{source: ChangeDetailsState, err: fmt.Errorf("unsupported change detail field: %s", field)}
		}
		change, err := client.GetChange(id)
		return changeSavedMsg{source: ChangeDetailsState, change: change, err: err}
	}
}

func changeDetailTextUpdateCommand(client appClient, source State, change dto.Change, field detailEditField, value string) tea.Cmd {
	return func() tea.Msg {
		id, err := changeNumericID(change)
		if err != nil {
			return changeSavedMsg{source: source, err: err}
		}
		switch field {
		case detailEditTitle:
			if _, err := client.UpdateChangeTitle(id, value); err != nil {
				return changeSavedMsg{source: source, err: err}
			}
		case detailEditRequirement:
			if _, err := client.UpdateChangeBody(id, value); err != nil {
				return changeSavedMsg{source: source, err: err}
			}
		case detailEditPullRequest:
			if _, err := client.UpdateChangePRBody(id, value); err != nil {
				return changeSavedMsg{source: source, err: err}
			}
		case detailEditPRUrl:
			if _, err := client.UpdateChangePRUrl(id, value); err != nil {
				return changeSavedMsg{source: source, err: err}
			}
		default:
			return changeSavedMsg{source: source, err: fmt.Errorf("unsupported change detail text field: %s", field)}
		}
		change, err := client.GetChange(id)
		return changeSavedMsg{source: source, change: change, err: err}
	}
}

func changeReferenceData(client appClient, projectID string, body string) ([]dto.Option, []dto.Option, error) {
	types, err := client.ListTypes()
	if err != nil {
		return nil, nil, err
	}
	if changes.RequirementEpicName(body) == "" {
		return types, nil, nil
	}
	epics, err := client.ListEpics(projectID)
	if err != nil {
		return nil, nil, err
	}
	return types, epics, nil
}

func changeNumericID(change dto.Change) (int, error) {
	id, err := strconv.Atoi(strings.TrimSpace(change.ID))
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("change ID must be a valid positive number")
	}
	return id, nil
}

func currentProjectNumericID(projectID string) (int, error) {
	id, err := strconv.Atoi(strings.TrimSpace(projectID))
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("current project must be numeric")
	}
	return id, nil
}

func sameEpicID(parsed *int, original string) bool {
	original = strings.TrimSpace(original)
	if parsed == nil {
		return original == ""
	}
	return original == strconv.Itoa(*parsed)
}

func toggleChangeType(current []string, selected dto.Option) []string {
	selectedID := strings.TrimSpace(selected.ID)
	if selectedID == "" {
		selectedID = strings.TrimSpace(selected.Label)
	}
	next := make([]string, 0, len(current)+1)
	removed := false
	for _, changeType := range current {
		if changeType == selectedID || changeType == selected.Label {
			removed = true
			continue
		}
		next = append(next, changeType)
	}
	if !removed && selectedID != "" {
		next = append(next, selectedID)
	}
	sort.Strings(next)
	return next
}

func selectedEpicID(selected dto.Option) (*int, error) {
	if selected.ID == "@none" {
		return nil, nil
	}
	epicID, err := strconv.Atoi(strings.TrimSpace(selected.ID))
	if err != nil || epicID <= 0 {
		return nil, fmt.Errorf("epic ID must be numeric")
	}
	return &epicID, nil
}
