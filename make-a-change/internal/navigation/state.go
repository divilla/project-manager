package navigation

// State identifies an mch screen, overlay, or confirmation state.
type State string

// State constants name every screen and overlay in the navigation shell.
const (
	MainState                     State = "MainState"
	ChangesListState              State = "ChangesListState"
	ChangeDetailsState            State = "ChangeDetailsState"
	RequirementDetailsState       State = "RequirementDetailsState"
	ChangeCreateState             State = "ChangeCreateState"
	ChangeUpdateState             State = "ChangeUpdateState"
	RequirementCreateState        State = "RequirementCreateState"
	RequirementUpdateState        State = "RequirementUpdateState"
	EpicsListState                State = "EpicsListState"
	EpicDetailsState              State = "EpicDetailsState"
	EpicCreateState               State = "EpicCreateState"
	EpicUpdateState               State = "EpicUpdateState"
	ProjectsListState             State = "ProjectsListState"
	ProjectDetailsState           State = "ProjectDetailsState"
	ProjectCreateState            State = "ProjectCreateState"
	ProjectUpdateState            State = "ProjectUpdateState"
	MainHelpState                 State = "MainHelpState"
	ChangesHelpState              State = "ChangesHelpState"
	EpicsHelpState                State = "EpicsHelpState"
	ProjectsHelpState             State = "ProjectsHelpState"
	FindInputState                State = "FindInput"
	CommandDropDownState          State = "CommandDropDown"
	ListSelectionDropDownState    State = "ListSelectionDropDown"
	SelectProjectDropDown         State = "SelectProjectDropDown"
	SelectPhaseDropDown           State = "SelectPhaseDropDown"
	SelectEpicDropDown            State = "SelectEpicDropDown"
	SelectTypesDropDown           State = "SelectTypesDropDown"
	ChangeDeleteConfirmation      State = "ChangeDeleteConfirmation"
	RequirementDeleteConfirmation State = "RequirementDeleteConfirmation"
	EpicDeleteConfirmation        State = "EpicDeleteConfirmation"
	ProjectDeleteConfirmation     State = "ProjectDeleteConfirmation"
	DoneState                     State = "DoneState"
)
