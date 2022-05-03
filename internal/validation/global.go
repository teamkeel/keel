package validation

type globalValidationContext struct {
	modelNamesUsed map[string]bool
}

func newGlobalContext() *globalValidationContext {
	return &globalValidationContext{
		modelNamesUsed: map[string]bool{},
	}
}