package changeview

import (
	"aipm/internal/dto"
	"aipm/pkg/markdown"
)

// ChangeRenderer defines ChangeRenderer values.
type ChangeRenderer struct {
	parser    markdown.Parser
	sanitizer markdown.Sanitizer
}

// NewChangeRenderer initializes or executes NewChangeRenderer behavior.
func NewChangeRenderer(parser markdown.Parser, sanitizer markdown.Sanitizer) ChangeRenderer {
	return ChangeRenderer{parser: parser, sanitizer: sanitizer}
}

// RenderChange executes RenderChange behavior.
func (r ChangeRenderer) RenderChange(change dto.Change) dto.Change {
	if r.parser == nil || r.sanitizer == nil || change.Body == "" {
		return change
	}
	change.BodyHTML = r.sanitizer.Parse(r.parser.Parse(change.Body))
	return change
}

// RenderMutation executes RenderMutation behavior.
func (r ChangeRenderer) RenderMutation(mutation dto.RequirementMutationResponse) dto.RequirementMutationResponse {
	mutation.Change = r.RenderChange(mutation.Change)
	return mutation
}
