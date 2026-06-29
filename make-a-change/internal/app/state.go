package app

import "mch/internal/navigation"

// State identifies the active mch screen or interaction state.
type State = navigation.State

// State aliases keep app tests and callers decoupled from the navigation package path.
const (
	MainState                     = navigation.MainState
	ChangesListState              = navigation.ChangesListState
	ChangeDetailsState            = navigation.ChangeDetailsState
	RequirementDetailsState       = navigation.RequirementDetailsState
	ChangeCreateState             = navigation.ChangeCreateState
	ChangeUpdateState             = navigation.ChangeUpdateState
	RequirementCreateState        = navigation.RequirementCreateState
	RequirementUpdateState        = navigation.RequirementUpdateState
	EpicsListState                = navigation.EpicsListState
	EpicDetailsState              = navigation.EpicDetailsState
	EpicCreateState               = navigation.EpicCreateState
	EpicUpdateState               = navigation.EpicUpdateState
	ProjectsListState             = navigation.ProjectsListState
	ProjectDetailsState           = navigation.ProjectDetailsState
	ProjectCreateState            = navigation.ProjectCreateState
	ProjectUpdateState            = navigation.ProjectUpdateState
	MainHelpState                 = navigation.MainHelpState
	ChangesHelpState              = navigation.ChangesHelpState
	EpicsHelpState                = navigation.EpicsHelpState
	ProjectsHelpState             = navigation.ProjectsHelpState
	FindInputState                = navigation.FindInputState
	CommandDropDownState          = navigation.CommandDropDownState
	ListSelectionDropDownState    = navigation.ListSelectionDropDownState
	SelectProjectDropDown         = navigation.SelectProjectDropDown
	SelectPhaseDropDown           = navigation.SelectPhaseDropDown
	SelectEpicDropDown            = navigation.SelectEpicDropDown
	SelectTypesDropDown           = navigation.SelectTypesDropDown
	ChangeDeleteConfirmation      = navigation.ChangeDeleteConfirmation
	RequirementDeleteConfirmation = navigation.RequirementDeleteConfirmation
	EpicDeleteConfirmation        = navigation.EpicDeleteConfirmation
	ProjectDeleteConfirmation     = navigation.ProjectDeleteConfirmation
	DoneState                     = navigation.DoneState
)
