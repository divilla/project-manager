package execution

import (
	"apih/skeleton/internal/domain"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
)

var ErrInvalidDirectoryTree = errors.New("invalid directory tree")

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
	dirs, err := collectDirs(suite)
	if err != nil {
		return 1, err
	}

	if err := executeStages(ctx, dirs, s.processDir); err != nil {
		return 1, err
	}

	return 0, nil
}

func collectDirs(suite *domain.Suite) ([][]*domain.Directory, error) {
	if suite == nil {
		return nil, fmt.Errorf("%w: suite is nil", ErrInvalidDirectoryTree)
	}
	if suite.Root == nil {
		return nil, fmt.Errorf("%w: root is nil", ErrInvalidDirectoryTree)
	}
	if suite.Root.Stage != 0 {
		return nil, fmt.Errorf("%w: root stage is %d, expected 0", ErrInvalidDirectoryTree, suite.Root.Stage)
	}

	dirs := make([][]*domain.Directory, 1)
	seen := make(map[*domain.Directory]struct{})

	var visit func(*domain.Directory, *domain.Directory) error
	visit = func(dir, parent *domain.Directory) error {
		if dir == nil {
			return fmt.Errorf("%w: nil child", ErrInvalidDirectoryTree)
		}
		if _, ok := seen[dir]; ok {
			return fmt.Errorf("%w: repeated directory %q", ErrInvalidDirectoryTree, dir.Path)
		}
		seen[dir] = struct{}{}

		if parent != nil {
			if dir.Parent != parent {
				return fmt.Errorf("%w: directory %q has an invalid parent", ErrInvalidDirectoryTree, dir.Path)
			}
			if dir.Stage != parent.Stage+1 {
				return fmt.Errorf(
					"%w: directory %q has stage %d, expected %d",
					ErrInvalidDirectoryTree,
					dir.Path,
					dir.Stage,
					parent.Stage+1,
				)
			}
		}

		for len(dirs) <= dir.Stage {
			dirs = append(dirs, nil)
		}
		dirs[dir.Stage] = append(dirs[dir.Stage], dir)

		for _, child := range dir.Children {
			if err := visit(child, dir); err != nil {
				return err
			}
		}

		return nil
	}

	if err := visit(suite.Root, nil); err != nil {
		return nil, err
	}

	return dirs, nil
}

type directoryProcessor func(context.Context, *domain.Directory) error

func executeStages(
	ctx context.Context,
	dirs [][]*domain.Directory,
	process directoryProcessor,
) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, stage := range dirs {
		if err := executeStage(ctx, cancel, stage, process); err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
	}

	return nil
}

func executeStage(
	ctx context.Context,
	cancel context.CancelFunc,
	dirs []*domain.Directory,
	process directoryProcessor,
) error {
	var wg sync.WaitGroup
	var firstErr error
	var firstErrOnce sync.Once

	for _, dir := range dirs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := process(ctx, dir); err != nil {
				firstErrOnce.Do(func() {
					firstErr = err
					cancel()
				})
			}
		}()
	}

	wg.Wait()
	return firstErr
}

func (s *StepRunner) processDir(ctx context.Context, dir *domain.Directory) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	_ = s
	_ = dir
	return nil
}
