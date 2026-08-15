package orchestrator_mock

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"
)

const (
	ExitSuccess          = 0
	ExitValidationFailed = 1
	ExitConfiguration    = 2
	ExitMissingTool      = 3
	ExitInternal         = 4
)

// OrchestratorMock owns the complete application flow, including cancellation,
// stage barriers, concurrent files, sequential steps within a file, and output.
// Service size: BIG (approximately 700-900 production lines).
type OrchestratorMock struct {
	services Services
	now      func() time.Time
}

func New(services Services) *OrchestratorMock {
	return &OrchestratorMock{services: services, now: time.Now}
}

// Run executes the mocked service graph. Configuration-phase errors use exit
// code 2, preflight errors use 3, and execution/output errors use 4. A mock may
// preserve an external tool's exact exit status by returning an ExitError.
func (o *OrchestratorMock) Run(ctx context.Context, options RunOptions) RunResult {
	started := o.now()
	if err := o.emit(ctx, Event{Kind: EventSuiteStarted, Timestamp: started}); err != nil {
		return o.failed(SuiteResult{}, err, ExitInternal)
	}

	files, err := o.services.Loader.Files(ctx, options.WorkDir)
	if err != nil {
		return o.failed(SuiteResult{}, err, ExitConfiguration)
	}
	suiteTree, err := o.services.Parser.Parse(ctx, options.WorkDir, files)
	if err != nil {
		return o.failed(SuiteResult{}, err, ExitConfiguration)
	}
	suiteTree, err = o.services.Selector.Select(ctx, suiteTree, options.Filter)
	if err != nil {
		return o.failed(SuiteResult{}, err, ExitConfiguration)
	}
	if err := o.services.Defaults.Resolve(ctx, suiteTree); err != nil {
		return o.failed(SuiteResult{}, err, ExitConfiguration)
	}
	if err := o.services.Steps.Resolve(ctx, suiteTree); err != nil {
		return o.failed(SuiteResult{}, err, ExitConfiguration)
	}
	stages, err := o.services.Stages.Plan(ctx, suiteTree)
	if err != nil {
		return o.failed(SuiteResult{}, err, ExitConfiguration)
	}
	sort.SliceStable(stages, func(i, j int) bool { return stages[i].Depth < stages[j].Depth })
	if err := o.services.Tools.Check(ctx, stages); err != nil {
		return o.failed(SuiteResult{}, err, ExitMissingTool)
	}

	suite := SuiteResult{Summary: Summary{Started: started}}
	for _, stage := range stages {
		results, runErr := o.runStage(ctx, stage)
		suite.Steps = append(suite.Steps, results...)
		if runErr != nil {
			o.finishSummary(&suite)
			return o.failed(suite, runErr, ExitInternal)
		}
	}

	o.finishSummary(&suite)
	exitCode := ExitSuccess
	if suite.Summary.Failed > 0 {
		exitCode = ExitValidationFailed
	}
	if err := o.emit(ctx, Event{Kind: EventSummary, Timestamp: suite.Summary.Ended, Summary: &suite.Summary}); err != nil {
		return o.failed(suite, err, ExitInternal)
	}
	return RunResult{Suite: suite, ExitCode: exitCode}
}

type fileResult struct {
	results []StepResult
	err     error
}

func (o *OrchestratorMock) runStage(parent context.Context, stage Stage) ([]StepResult, error) {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	files := stageFiles(stage)
	completed := make(chan fileResult, len(files))
	var workers sync.WaitGroup
	for _, file := range files {
		file := file
		workers.Add(1)
		go func() {
			defer workers.Done()
			completed <- o.runFile(ctx, file)
		}()
	}
	go func() {
		workers.Wait()
		close(completed)
	}()

	var results []StepResult
	var firstErr error
	for completedFile := range completed {
		results = append(results, completedFile.results...)
		if completedFile.err != nil && firstErr == nil {
			firstErr = completedFile.err
			cancel()
		}
	}
	return results, firstErr
}

func (o *OrchestratorMock) runFile(ctx context.Context, file *File) fileResult {
	var results []StepResult
	for _, step := range file.RuntimeSteps {
		if err := ctx.Err(); err != nil {
			return fileResult{results: results, err: err}
		}
		result, err := o.services.Runner.Run(ctx, step)
		if err != nil {
			return fileResult{results: results, err: err}
		}
		results = append(results, result)
		resultCopy := result
		if err := o.emit(ctx, Event{
			Kind:      EventStepFinished,
			Timestamp: o.now(),
			Step:      &resultCopy,
		}); err != nil {
			return fileResult{results: results, err: err}
		}
	}
	return fileResult{results: results}
}

func stageFiles(stage Stage) []*File {
	var files []*File
	for _, directory := range stage.Directories {
		if directory == nil {
			continue
		}
		for _, file := range directory.Files {
			if file != nil && len(file.RuntimeSteps) > 0 {
				files = append(files, file)
			}
		}
	}
	return files
}

func (o *OrchestratorMock) finishSummary(suite *SuiteResult) {
	suite.Summary.Total = len(suite.Steps)
	for _, result := range suite.Steps {
		if result.Passed() {
			suite.Summary.Passed++
		} else {
			suite.Summary.Failed++
		}
	}
	suite.Summary.Ended = o.now()
}

func (o *OrchestratorMock) emit(ctx context.Context, event Event) error {
	return o.services.Output.Emit(ctx, event)
}

func (o *OrchestratorMock) failed(suite SuiteResult, err error, fallback int) RunResult {
	exitCode := fallback
	var exitErr interface{ ExitCode() int }
	if errors.As(err, &exitErr) {
		exitCode = exitErr.ExitCode()
	}
	return RunResult{Suite: suite, ExitCode: exitCode, Err: err}
}

// ExitError lets tool mocks preserve curl, jq, or Git exit statuses.
type ExitError struct {
	Code  int
	Cause error
}

func (e *ExitError) Error() string {
	if e.Cause == nil {
		return "external tool failed"
	}
	return e.Cause.Error()
}

func (e *ExitError) Unwrap() error { return e.Cause }

func (e *ExitError) ExitCode() int { return e.Code }
