package changeview

import (
	"aipm/internal/dto"
	"aipm/pkg/markdown"
)

type ChangeRenderer struct {
	parser    markdown.Parser
	sanitizer markdown.Sanitizer
}

func NewChangeRenderer(parser markdown.Parser, sanitizer markdown.Sanitizer) ChangeRenderer {
	return ChangeRenderer{parser: parser, sanitizer: sanitizer}
}

func (r ChangeRenderer) RenderChange(change dto.Change) dto.Change {
	if r.parser == nil || r.sanitizer == nil || change.Body == "" {
		return change
	}
	change.BodyHTML = r.sanitizer.Parse(r.parser.Parse(change.Body))
	return change
}

func (r ChangeRenderer) RenderMutation(mutation dto.RequirementMutationResponse) dto.RequirementMutationResponse {
	mutation.Change = r.RenderChange(mutation.Change)
	return mutation
}
