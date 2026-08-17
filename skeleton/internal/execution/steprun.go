package execution

import (
	"apih/skeleton/internal/domain"
	"context"
	"io"
	"sync"
)

type StepRunner struct {
	varProc *VariableProcessor
	val     *Validator
	out     io.Writer
}

func NewStepRunner(
	variableProcessor *VariableProcessor,
	validator *Validator,
	output io.Writer,
) *StepRunner {
	return &StepRunner{
		varProc: variableProcessor,
		val:     validator,
		out:     output,
	}
}

// Prepare traverses directories through children, starting from suite.Root
// For each directory it iterates directory.ResolvedSteps and executes varProc.Load and varProc.ParseRequestBody
func (s *StepRunner) Prepare(
	ctx context.Context,
	suite *domain.Suite,
) error {
	return nil
}

// Execute traverses directories through children, starting from suite.Root, from goroutines, one for each directory,
// until entire same number stage is executed. For each directory it iterates directory.ResolvedSteps and executes
// runner.Curl, varProc.ParseResponseExpected, val.ValidateTypes, val.ValidateExpected and finally varProc.Capture
// On detected validation error, Execute does not return error, but reports failed validation to standard s.out.
// Once it finishes traversing in one or more validation failed it will return exit code 1, ValidationError error
func (s *StepRunner) Execute(
	ctx context.Context,
	suite *domain.Suite,
) (int, error) {
	dirs := make([][]*domain.Directory, 255)
	dirs = fillDirs(dirs, suite.Root)

	for _, ds := range dirs {
		wg := sync.WaitGroup{}
		for _, dir := range ds {
			wg.Add(1)
			go processDir(&wg, dir)
		}
		wg.Wait()
	}

	return 0, nil
}

func fillDirs(dirs [][]*domain.Directory, dir *domain.Directory) [][]*domain.Directory {
	dirs[dir.Stage] = append(dirs[dir.Stage], dir)
	for _, d := range dir.Children {
		dirs = fillDirs(dirs, d)
	}

	return dirs
}

func processDir(wg *sync.WaitGroup, dir *domain.Directory) {
	defer wg.Done()
	_ = dir
}
