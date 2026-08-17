package errs

import (
	"apih/skeleton/internal/domain"
	"errors"
)

func DefaultsDefinitionError(defaults *domain.DefaultsDefinition, yamlPath string, errStatic, errOriginal error) error {
	return errors.New("defaults error in file [relative-yaml-file-path]:[line-number]: [yaml-path]: [errStatic]: [optional-errOriginal]")
}

func StepDefinitionError(step *domain.StepsDefinition, yamlPath string, errStatic, errOriginal error) error {
	return errors.New("definition error in file [relative-yaml-file-path]:[line-number]: [yaml-path]: [errStatic]: [optional-errOriginal]")
}

func StepExecutionError(step *domain.Step, yamlPath string, errStatic, errOriginal error) error {
	return errors.New("execution error in file [relative-yaml-file-path]:[line-number]: [yaml-path]: [errStatic]: [optional-errOriginal]")
}
