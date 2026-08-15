package orchestrator_mock

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
)

func TestOrchestratorRunsStagesInOrderAndStepsInAFileSequentially(t *testing.T) {
	root := &Directory{Depth: 0, Dir: "suite"}
	suiteTree := &Suite{WorkDir: "suite", Root: root}
	first := &File{RuntimeSteps: []RuntimeStep{{ID: "one"}, {ID: "two"}}}
	second := &File{RuntimeSteps: []RuntimeStep{{ID: "three"}}}
	stages := []Stage{
		{Depth: 1, Directories: []*Directory{{Files: []*File{second}}}},
		{Depth: 0, Directories: []*Directory{{Files: []*File{first}}}},
	}

	var mu sync.Mutex
	var calls []string
	services := successfulServices(suiteTree, stages)
	services.Runner.RunFunc = func(_ context.Context, step RuntimeStep) (StepResult, error) {
		mu.Lock()
		calls = append(calls, step.ID)
		mu.Unlock()
		return StepResult{Step: step}, nil
	}

	got := New(services).Run(context.Background(), RunOptions{WorkDir: "suite"})
	if got.Err != nil {
		t.Fatalf("Run() error = %v", got.Err)
	}
	if got.ExitCode != ExitSuccess {
		t.Fatalf("Run() exit code = %d, want %d", got.ExitCode, ExitSuccess)
	}
	if !reflect.DeepEqual(calls, []string{"one", "two", "three"}) {
		t.Fatalf("runner calls = %v, want [one two three]", calls)
	}
	if got.Suite.Summary.Total != 3 || got.Suite.Summary.Passed != 3 {
		t.Fatalf("summary = %+v, want three passing steps", got.Suite.Summary)
	}
}

func TestOrchestratorCollectsValidationFailuresAndContinues(t *testing.T) {
	root := &Directory{Depth: 0, Dir: "suite"}
	suiteTree := &Suite{WorkDir: "suite", Root: root}
	file := &File{RuntimeSteps: []RuntimeStep{{ID: "fails"}, {ID: "passes"}}}
	stages := []Stage{{Depth: 0, Directories: []*Directory{{Files: []*File{file}}}}}
	services := successfulServices(suiteTree, stages)
	services.Runner.RunFunc = func(_ context.Context, step RuntimeStep) (StepResult, error) {
		result := StepResult{Step: step}
		if step.ID == "fails" {
			result.Failures = []ValidationFailure{{Kind: FailureStatus, Message: "unexpected status"}}
		}
		return result, nil
	}

	got := New(services).Run(context.Background(), RunOptions{WorkDir: "suite"})
	if got.Err != nil {
		t.Fatalf("Run() error = %v", got.Err)
	}
	if got.ExitCode != ExitValidationFailed {
		t.Fatalf("Run() exit code = %d, want %d", got.ExitCode, ExitValidationFailed)
	}
	if got.Suite.Summary.Total != 2 || got.Suite.Summary.Passed != 1 || got.Suite.Summary.Failed != 1 {
		t.Fatalf("summary = %+v, want one pass and one failure", got.Suite.Summary)
	}
}

func TestOrchestratorPreservesExternalToolExitCode(t *testing.T) {
	root := &Directory{Depth: 0, Dir: "suite"}
	suiteTree := &Suite{WorkDir: "suite", Root: root}
	file := &File{RuntimeSteps: []RuntimeStep{{ID: "tool-failure"}}}
	stages := []Stage{{Depth: 0, Directories: []*Directory{{Files: []*File{file}}}}}
	services := successfulServices(suiteTree, stages)
	services.Runner.RunFunc = func(context.Context, RuntimeStep) (StepResult, error) {
		return StepResult{}, &ExitError{Code: 17, Cause: errors.New("jq failed")}
	}

	got := New(services).Run(context.Background(), RunOptions{WorkDir: "suite"})
	if got.ExitCode != 17 {
		t.Fatalf("Run() exit code = %d, want 17", got.ExitCode)
	}
	if got.Err == nil || got.Err.Error() != "jq failed" {
		t.Fatalf("Run() error = %v, want jq failure", got.Err)
	}
}

func successfulServices(suiteTree *Suite, stages []Stage) Services {
	return Services{
		Loader: &LoaderMock{FilesFunc: func(context.Context, string) ([]string, error) {
			return []string{"suite/root.yaml"}, nil
		}},
		Parser: &ParserMock{ParseFunc: func(context.Context, string, []string) (*Suite, error) {
			return suiteTree, nil
		}},
		Selector: &SelectorMock{SelectFunc: func(_ context.Context, got *Suite, _ Filter) (*Suite, error) {
			return got, nil
		}},
		Defaults: &DefaultsResolverMock{ResolveFunc: func(context.Context, *Suite) error {
			return nil
		}},
		Steps: &StepResolverMock{ResolveFunc: func(context.Context, *Suite) error {
			return nil
		}},
		Stages: &StagePlannerMock{PlanFunc: func(context.Context, *Suite) ([]Stage, error) {
			return stages, nil
		}},
		Tools: &ToolPreflightMock{CheckFunc: func(context.Context, []Stage) error {
			return nil
		}},
		Runner: &StepRunnerMock{RunFunc: func(_ context.Context, step RuntimeStep) (StepResult, error) {
			return StepResult{Step: step}, nil
		}},
		Output: &OutputMock{EmitFunc: func(context.Context, Event) error {
			return nil
		}},
	}
}
