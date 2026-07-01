package app

import (
	"fmt"
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
		if _, err := changes.ParseRequirementBodyStructure(body); err != nil {
			return changeSavedMsg{source: ChangeCreateState, err: err}
		}
		types, epics, err := changeReferenceData(client, projectIDValue, body)
		if err != nil {
			return changeSavedMsg{source: ChangeCreateState, err: err}
		}
		parsed, err := changes.ParseRequirementBody(body, types, epics)
		if err != nil {
			return changeSavedMsg{source: ChangeCreateState, err: err}
		}
		created, err := client.CreateChange(dto.ChangeCreateInput{
			ProjectID:       projectID,
			Title:           parsed.Title,
			RequirementBody: parsed.RequirementBody,
			ChangeTypes:     parsed.ChangeTypes,
			EpicID:          parsed.EpicID,
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
		if _, err := changes.ParseRequirementBodyStructure(body); err != nil {
			return changeSavedMsg{source: ChangeUpdateState, err: err}
		}
		types, epics, err := changeReferenceData(client, projectID, body)
		if err != nil {
			return changeSavedMsg{source: ChangeUpdateState, err: err}
		}
		parsed, err := changes.ParseRequirementBody(body, types, epics)
		if err != nil {
			return changeSavedMsg{source: ChangeUpdateState, err: err}
		}
		if parsed.Title != original.Title {
			if _, err := client.UpdateChangeTitle(id, parsed.Title); err != nil {
				return changeSavedMsg{source: ChangeUpdateState, err: err}
			}
		}
		if parsed.RequirementBody != original.RequirementBody {
			if _, err := client.UpdateChangeRequirementBody(id, parsed.RequirementBody); err != nil {
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
