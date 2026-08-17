package reporter

import (
	"apih/skeleton/internal/domain"
	"apih/skeleton/pkg/errs"
	"context"
	"errors"
	"fmt"
	"io"
)

var ReporterError = errors.New("reporter error")

// Reporter owns all human-readable standard output. The writer is normally
// os.Stdout and may be replaced by a buffer or another writer in tests.
type Reporter struct {
	output io.Writer
}

func NewReporter(output io.Writer) *Reporter {
	return &Reporter{output: output}
}

func (r *Reporter) WorkingDirectory(workDir string) error {
	if r == nil || r.output == nil {
		return errs.Build(errs.ExitInternal, ReporterError, nil, "output is nil")
	}
	if _, err := fmt.Fprintf(r.output, "Working Directory: %s\n\n", workDir); err != nil {
		return errs.Build(errs.ExitInternal, ReporterError, err)
	}
	return nil
}

// Success reports a directory whose execution completed without validation
// failures. Formatting is intentionally left to the Reporter implementation.
func (r *Reporter) Success(ctx context.Context, directory *domain.Directory) error {
	return nil
}

// FailureTypes reports one nonfatal response-type validation failure.
func (r *Reporter) FailureTypes(ctx context.Context, step *domain.Step, failure error) error {
	return nil
}

// FailureExpected reports one nonfatal expected-response validation failure.
func (r *Reporter) FailureExpected(ctx context.Context, step *domain.Step, failure error) error {
	return nil
}

// Debug reports the final runtime state of a selected debug step.
func (r *Reporter) Debug(ctx context.Context, step *domain.Step) error {
	return nil
}
