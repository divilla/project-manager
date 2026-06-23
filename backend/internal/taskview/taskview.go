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

func (r TaskRenderer) RenderTask(task dto.Task) dto.Task {
	if r.parser == nil || r.sanitizer == nil || task.Description == "" {
		return task
	}
	task.DescriptionHTML = r.sanitizer.Parse(r.parser.Parse(task.Description))
	return task
}

func (r TaskRenderer) RenderMutation(mutation dto.RequirementMutationResponse) dto.RequirementMutationResponse {
	mutation.Task = r.RenderTask(mutation.Task)
	return mutation
}
