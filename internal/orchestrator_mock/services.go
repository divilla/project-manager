package orchestrator_mock

import (
	"context"
	"errors"
)

var ErrUnconfiguredMock = errors.New("orchestrator mock service is not configured")

// LoaderMock models internal/suite discovery.
// Service size: SMALL (approximately 80-150 production lines).
type LoaderMock struct {
	FilesFunc func(context.Context, string) ([]string, error)
}

func (m *LoaderMock) Files(ctx context.Context, workDir string) ([]string, error) {
	if m == nil || m.FilesFunc == nil {
		return nil, ErrUnconfiguredMock
	}
	return m.FilesFunc(ctx, workDir)
}

// ParserMock models strict YAML decoding, duplicate-key checks, and SourceRefs.
// Service size: BIG (approximately 700-950 production lines).
type ParserMock struct {
	ParseFunc func(context.Context, string, []string) (*Suite, error)
}

func (m *ParserMock) Parse(ctx context.Context, workDir string, files []string) (*Suite, error) {
	if m == nil || m.ParseFunc == nil {
		return nil, ErrUnconfiguredMock
	}
	return m.ParseFunc(ctx, workDir, files)
}

// SelectorMock models metadata name/label filtering.
// Service size: SMALL (approximately 100-180 production lines).
type SelectorMock struct {
	SelectFunc func(context.Context, *Suite, Filter) (*Suite, error)
}

func (m *SelectorMock) Select(ctx context.Context, suite *Suite, filter Filter) (*Suite, error) {
	if m == nil || m.SelectFunc == nil {
		return nil, ErrUnconfiguredMock
	}
	return m.SelectFunc(ctx, suite, filter)
}

// DefaultsResolverMock models parent discovery, inheritance, and headers.
// Service size: MEDIUM (approximately 300-450 production lines).
type DefaultsResolverMock struct {
	ResolveFunc func(context.Context, *Suite) error
}

func (m *DefaultsResolverMock) Resolve(ctx context.Context, suite *Suite) error {
	if m == nil || m.ResolveFunc == nil {
		return ErrUnconfiguredMock
	}
	return m.ResolveFunc(ctx, suite)
}

// StepResolverMock models defaults overlays, URL construction, and
// pre-execution validation.
// Service size: MEDIUM (approximately 400-600 production lines).
type StepResolverMock struct {
	ResolveFunc func(context.Context, *Suite) error
}

func (m *StepResolverMock) Resolve(ctx context.Context, suite *Suite) error {
	if m == nil || m.ResolveFunc == nil {
		return ErrUnconfiguredMock
	}
	return m.ResolveFunc(ctx, suite)
}

// StagePlannerMock models grouping directories by relative depth.
// Service size: SMALL (approximately 100-160 production lines).
type StagePlannerMock struct {
	PlanFunc func(context.Context, *Suite) ([]Stage, error)
}

func (m *StagePlannerMock) Plan(ctx context.Context, suite *Suite) ([]Stage, error) {
	if m == nil || m.PlanFunc == nil {
		return nil, ErrUnconfiguredMock
	}
	return m.PlanFunc(ctx, suite)
}

// ToolPreflightMock models required-tool derivation and PATH checks.
// Service size: SMALL (approximately 100-180 production lines).
type ToolPreflightMock struct {
	CheckFunc func(context.Context, []Stage) error
}

func (m *ToolPreflightMock) Check(ctx context.Context, stages []Stage) error {
	if m == nil || m.CheckFunc == nil {
		return ErrUnconfiguredMock
	}
	return m.CheckFunc(ctx, stages)
}

// VariableStoreMock models the concurrent, write-once run-wide variable store.
// Service size: MEDIUM (approximately 220-320 production lines).
type VariableStoreMock struct {
	SetFunc func(context.Context, string, string, SourceRef) error
	GetFunc func(context.Context, string, SourceRef) (string, error)
}

func (m *VariableStoreMock) Set(ctx context.Context, key, value string, source SourceRef) error {
	if m == nil || m.SetFunc == nil {
		return ErrUnconfiguredMock
	}
	return m.SetFunc(ctx, key, value, source)
}

func (m *VariableStoreMock) Get(ctx context.Context, key string, source SourceRef) (string, error) {
	if m == nil || m.GetFunc == nil {
		return "", ErrUnconfiguredMock
	}
	return m.GetFunc(ctx, key, source)
}

// SubstitutorMock models whole-value and embedded variable expansion.
// Service size: MEDIUM (approximately 220-350 production lines).
type SubstitutorMock struct {
	SubstituteFunc func(context.Context, YAMLString, SourceRef) (YAMLString, error)
}

func (m *SubstitutorMock) Substitute(ctx context.Context, value YAMLString, source SourceRef) (YAMLString, error) {
	if m == nil || m.SubstituteFunc == nil {
		return "", ErrUnconfiguredMock
	}
	return m.SubstituteFunc(ctx, value, source)
}

// RequestExecutorMock models request preparation plus curl execution.
// Service size: MEDIUM (approximately 400-600 production lines).
type RequestExecutorMock struct {
	ExecuteFunc func(context.Context, RuntimeStep) (HTTPResponse, error)
}

func (m *RequestExecutorMock) Execute(ctx context.Context, step RuntimeStep) (HTTPResponse, error) {
	if m == nil || m.ExecuteFunc == nil {
		return HTTPResponse{}, ErrUnconfiguredMock
	}
	return m.ExecuteFunc(ctx, step)
}

// ResponseProcessorMock models JSON checks, capture, expected comparisons, and
// type assertions.
// Service size: BIG (approximately 750-1,000 production lines).
type ResponseProcessorMock struct {
	ProcessFunc func(context.Context, RuntimeStep, HTTPResponse) (map[string]string, []ValidationFailure, error)
}

func (m *ResponseProcessorMock) Process(ctx context.Context, step RuntimeStep, response HTTPResponse) (map[string]string, []ValidationFailure, error) {
	if m == nil || m.ProcessFunc == nil {
		return nil, nil, ErrUnconfiguredMock
	}
	return m.ProcessFunc(ctx, step, response)
}

// StepRunnerMock is the seam used by the orchestrator. Its production version
// composes the variable, substitution, request, and response services above.
// Service size: MEDIUM (approximately 300-500 production lines).
type StepRunnerMock struct {
	RunFunc func(context.Context, RuntimeStep) (StepResult, error)
}

func (m *StepRunnerMock) Run(ctx context.Context, step RuntimeStep) (StepResult, error) {
	if m == nil || m.RunFunc == nil {
		return StepResult{}, ErrUnconfiguredMock
	}
	return m.RunFunc(ctx, step)
}

// OutputMock models terminal and streaming NDJSON presentation.
// Service size: MEDIUM (approximately 250-400 production lines).
type OutputMock struct {
	EmitFunc func(context.Context, Event) error
}

func (m *OutputMock) Emit(ctx context.Context, event Event) error {
	if m == nil || m.EmitFunc == nil {
		return ErrUnconfiguredMock
	}
	return m.EmitFunc(ctx, event)
}

// Services contains the service seams directly coordinated by Orchestrator.
// VariableStoreMock, SubstitutorMock, RequestExecutorMock, and
// ResponseProcessorMock are composed behind Runner in the production design.
type Services struct {
	Loader   *LoaderMock
	Parser   *ParserMock
	Selector *SelectorMock
	Defaults *DefaultsResolverMock
	Steps    *StepResolverMock
	Stages   *StagePlannerMock
	Tools    *ToolPreflightMock
	Runner   *StepRunnerMock
	Output   *OutputMock
}
