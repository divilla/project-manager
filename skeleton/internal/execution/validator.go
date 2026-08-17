package execution

import (
	"apih/skeleton/internal/domain"
	"context"
	"errors"
)

var ValidationError = errors.New("validation error")

type Validator struct{}

func (v *Validator) ValidateTypes(
	ctx context.Context,
	step *domain.Step,
) error {
	return nil
}

func (v *Validator) ValidateExpected(
	ctx context.Context,
	step *domain.Step,
) error {
	return nil
}
