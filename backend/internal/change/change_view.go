package change

import (
	"aipm/internal/dto"
	"aipm/pkg/markdown"
)

// Renderer defines Renderer values.
type Renderer struct {
	parser    markdown.Parser
	sanitizer markdown.Sanitizer
}

// NewRenderer initializes or executes NewRenderer behavior.
func NewRenderer(parser markdown.Parser, sanitizer markdown.Sanitizer) Renderer {
	return Renderer{parser: parser, sanitizer: sanitizer}
}

// RenderChange executes RenderChange behavior.
func (r Renderer) RenderChange(change dto.Change) dto.Change {
	if r.parser == nil || r.sanitizer == nil {
		return change
	}
	if change.Body != "" {
		change.HTML = r.sanitizer.Parse(r.parser.Parse(change.Body))
	}
	if change.PRBody != "" {
		change.PRHtml = r.sanitizer.Parse(r.parser.Parse(change.PRBody))
	}
	return change
}

// RenderMutation executes RenderMutation behavior.
func (r Renderer) RenderMutation(mutation dto.TestCaseMutationResponse) dto.TestCaseMutationResponse {
	mutation.Change = r.RenderChange(mutation.Change)
	return mutation
}
