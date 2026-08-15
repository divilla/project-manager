package orchestrator_mock

import "time"

// Value preserves the difference between an omitted value and an explicitly
// supplied zero value.
type Value[T any] struct {
	Value T
	Set   bool
}

type YAMLString string

type DocumentKind string

const (
	DocumentRoot     DocumentKind = "root"
	DocumentDefaults DocumentKind = "defaults"
	DocumentSteps    DocumentKind = "steps"
)

// SourceRef identifies the YAML node from which a value originated.
type SourceRef struct {
	FilePath string
	YAMLPath string
	Line     int
	Column   int
}

// SourceMap uses semantic paths such as request.body as keys.
type SourceMap map[string]SourceRef

type Metadata struct {
	Name   string
	Labels []string
}

// Defaults contains declarative values supplied by root and defaults documents.
// Value fields retain YAML presence so inheritance can distinguish omitted and
// explicitly empty values.
type Defaults struct {
	BaseURL  Value[string]
	BasePath Value[string]
	Headers  map[string]string
	Timeout  Value[int]
	Retries  Value[int]
	Sources  SourceMap
}

// RootFile is the mandatory marker and initial defaults source for a suite.
type RootFile struct {
	Metadata Metadata
	Defaults Defaults
	Sources  SourceMap
}

type DefaultsFile struct {
	Metadata Metadata
	Defaults Defaults
	Sources  SourceMap
}

type StepsFile struct {
	Metadata Metadata
	Steps    []Step
	Sources  SourceMap
}

type Request struct {
	Method   Value[string]
	BaseURL  Value[string]
	BasePath Value[string]
	Path     Value[string]
	Headers  map[string]string
	Timeout  Value[int]
	Retries  Value[int]
	Query    Value[string]
	Body     Value[YAMLString]
	Sources  SourceMap
}

type TypeDeclaration struct {
	BaseType string
	Zero     bool
	Optional bool
	Source   SourceRef
}

type ResponseExpectation struct {
	Status   Value[[]int]
	Capture  map[string]string
	Expected Value[YAMLString]
	Types    map[string]TypeDeclaration
	Sources  SourceMap
}

// Step is the declarative lifecycle state decoded from YAML.
type Step struct {
	Index    int
	Vars     map[string]any
	Request  Request
	Response *ResponseExpectation
	Sources  SourceMap
}

type Steps = []Step

// RuntimeDefaults is the effective inherited defaults for one directory.
type RuntimeDefaults struct {
	BaseURL  string
	BasePath Value[string]
	Headers  map[string]string
	Timeout  int
	Retries  int
	Sources  SourceMap
}

// RuntimeStep is a step after inherited defaults and available variables have
// been resolved. Body and Expected remain literal JSON text.
type RuntimeStep struct {
	ID          string
	FilePath    string
	Index       int
	Metadata    Metadata
	Vars        map[string]string
	Method      string
	URL         string
	Headers     map[string]string
	Timeout     int
	Retries     int
	Body        Value[YAMLString]
	Expectation *ResponseExpectation
	Sources     SourceMap
}

type RuntimeSteps = []RuntimeStep

// File carries the parsed and resolved state associated with one YAML file.
type File struct {
	FilePath        string
	Kind            DocumentKind
	Root            *RootFile
	Defaults        *DefaultsFile
	Steps           *StepsFile
	ParentDefaults  *Defaults
	RuntimeDefaults RuntimeDefaults
	RuntimeSteps    []RuntimeStep
	Sources         SourceMap
}

// Directory mirrors one node in the suite's filesystem hierarchy.
type Directory struct {
	Depth           int
	Dir             string
	RuntimeDefaults RuntimeDefaults
	Files           []*File
	Children        []*Directory
}

type Suite struct {
	WorkDir string
	Root    *Directory
}

// Stage is an execution barrier for directories at one relative depth.
type Stage struct {
	Depth       int
	Directories []*Directory
}

type Filter struct {
	Name   string
	Labels []string
}

type RunOptions struct {
	WorkDir string
	Filter  Filter
	JSON    bool
}

type Tool string

const (
	ToolCurl Tool = "curl"
	ToolJQ   Tool = "jq"
	ToolGit  Tool = "git"
)

type HTTPResponse struct {
	StatusCode int
	Headers    map[string][]string
	Body       []byte
}

type FailureKind string

const (
	FailureStatus   FailureKind = "status"
	FailureExpected FailureKind = "expected"
	FailureType     FailureKind = "type"
)

type ValidationFailure struct {
	Kind    FailureKind
	Message string
	Source  SourceRef
	Diff    string
}

// StepResult is the terminal lifecycle state for one RuntimeStep.
type StepResult struct {
	Step       RuntimeStep
	Response   HTTPResponse
	Captured   map[string]string
	Failures   []ValidationFailure
	StartedAt  time.Time
	FinishedAt time.Time
}

func (r StepResult) Passed() bool {
	return len(r.Failures) == 0
}

type Summary struct {
	Total   int
	Passed  int
	Failed  int
	Started time.Time
	Ended   time.Time
}

type SuiteResult struct {
	Steps   []StepResult
	Summary Summary
}

type EventKind string

const (
	EventSuiteStarted EventKind = "suite_started"
	EventStepFinished EventKind = "step_finished"
	EventError        EventKind = "error"
	EventSummary      EventKind = "summary"
)

// Event is shared by human and NDJSON output implementations.
type Event struct {
	Kind      EventKind
	Timestamp time.Time
	Step      *StepResult
	Summary   *Summary
	Error     *Diagnostic
}

type ErrorCategory string

const (
	ErrorConfiguration ErrorCategory = "configuration"
	ErrorResolution    ErrorCategory = "resolution"
	ErrorExecution     ErrorCategory = "execution"
	ErrorValidation    ErrorCategory = "validation"
	ErrorInternal      ErrorCategory = "internal"
)

// Diagnostic is a structured process-level error carrier.
type Diagnostic struct {
	Category ErrorCategory
	Message  string
	Source   SourceRef
	Tool     Tool
	ToolExit int
	Diff     string
}

func (d *Diagnostic) Error() string {
	if d == nil {
		return ""
	}
	return d.Message
}

func (d *Diagnostic) ExitCode() int {
	if d == nil {
		return ExitInternal
	}
	if d.ToolExit > 0 {
		return d.ToolExit
	}
	switch d.Category {
	case ErrorValidation:
		return ExitValidationFailed
	case ErrorConfiguration, ErrorResolution:
		return ExitConfiguration
	default:
		return ExitInternal
	}
}

type RunResult struct {
	Suite    SuiteResult
	ExitCode int
	Err      error
}
