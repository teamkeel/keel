package validation

import "github.com/teamkeel/keel/parser"

type Validator struct {
	schema *parser.Schema
}

func NewValidator(schema *parser.Schema) *Validator {
	return &Validator{
		schema: schema,
	}
}

func (vdr *Validator) RunAllValidators() error {
	validatorFuncs := []func(*parser.Schema) error {
		modelNames,
		fieldNames,
		etc,
	}
	for _, vf := range validatorFuncs {
		err := vf(vdr.schema)
		if err != nil {
			return err
		}
	}
	return nil
}

func modelNames(schema *parser.Schema) error {
	return nil
}

func fieldNames(schema *parser.Schema) error {
	return nil
}

func etc(schema *parser.Schema) error {
	return nil
}