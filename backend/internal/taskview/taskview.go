package taskview

import (
	"aipm/internal/dto"
	"aipm/pkg/markdown"
)

type TaskRenderer struct {
	parser    markdown.Parser
	sanitizer markdown.Sanitizer
}

func NewTaskRenderer(parser markdown.Parser, sanitizer markdown.Sanitizer) TaskRenderer {
	return TaskRenderer{parser: parser, sanitizer: sanitizer}
}

func (r TaskRenderer) RenderTask(task dto.Change) dto.Change {
	if r.parser == nil || r.sanitizer == nil || task.Body == "" {
		return task
	}
	task.BodyHTML = r.sanitizer.Parse(r.parser.Parse(task.Body))
	return task
}

func (r TaskRenderer) RenderMutation(mutation dto.RequirementMutationResponse) dto.RequirementMutationResponse {
	mutation.Change = r.RenderTask(mutation.Change)
	return mutation
}
