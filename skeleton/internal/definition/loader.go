package definition

import (
	"apih/skeleton/internal/domain"
	"context"
)

type Loader struct{}

func NewLoader() *Loader {
	return &Loader{}
}

// LoadDirectoryStructure traverses directory structure from suite.Workdir
// building *domain.Directory structure. *domain.Directory.Path is relative
// to suite.WorkDir, so suite.Directory.Path = "/"
func (l *Loader) LoadDirectoryStructure(
	ctx context.Context,
	suite *domain.Suite,
) error {
	return nil
}

// LoadDirectoryFiles traverses directory structure from suite.Directory
// populating each domain.Directory.Files with all directory *.yaml and *.yml files.
// LoadDirectoryFiles mutates only *domain.Directory.Files slice
func (l *Loader) LoadDirectoryFiles(
	ctx context.Context,
	suite *domain.Suite,
) error {
	return nil
}

// DecodeBaseDefinitions traverses directory structure from suite.Directory
// trying to decode suite.Directory.Files into domain.BaseDefinition.
// On success it updates file.Kind and populates suite.DefaultsDefinitions
// and suite.StepsDefinitions with
func (l *Loader) DecodeBaseDefinitions(
	ctx context.Context,
	suite *domain.Suite,
) error {
	return nil
}
