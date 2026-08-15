package suite

import "apih/internal/models"

type Parser struct {
	WorkDir string
	Files   []string
}

func NewParser(workDir string, files []string) *Parser {
	return &Parser{
		WorkDir: workDir,
		Files:   files,
	}
}

func (p *Parser) Parse() (*models.Suite, error) {
	return &models.Suite{WorkDir: p.WorkDir}, nil
}
