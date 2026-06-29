package changes

import "mch/internal/dto"

// Filters stores active change list filter selections.
type Filters struct {
	Phase dto.Option
	Epic  dto.Option
	Type  dto.Option
}
