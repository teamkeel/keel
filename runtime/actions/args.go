package actions

// GraphQL, RPC, and the Testing framework all have subtle differences in how arguments are processed
// (JSON from RPC requests or from the Testing framework use ISO8601 for instance, whereas GraphQL has its own special datetime format)
// This interface enforces a contract that all of these API implementations must use to normalize the params
// The end result is that the actions code is not aware of these differences, and deals with native Go entities like time.Time (in the case of Dates / Timestamps)
// In the future, there will be other types of data to account for where the input structure differs dependent on the source

type ArgParser interface {
	ParseGet(input map[string]any) (*Args, error)
	ParseCreate(input map[string]any) (*Args, error)
	ParseUpdate(input map[string]any) (*Args, error)
	ParseList(input map[string]any) (*Args, error)
	ParseDelete(input map[string]any) (*Args, error)
}

type ValueArgs map[string]any
type WhereArgs map[string]any

type Args struct {
	values ValueArgs
	wheres WhereArgs
}

func NewArgs(values ValueArgs, wheres WhereArgs) *Args {
	if values == nil || wheres == nil {
		panic("values or wheres input maps canot be nil in NewArgs")
	}

	return &Args{
		// Used to provide data for the means of writing (create and update)
		values: values,
		// Used to filter data before performing an action (get, list, update, delete)
		// Or inputs for use in expressions (i.e. explicit inputs)
		wheres: wheres,

		// TODO: ive realised that on some operations explicit inputs are passed to values and in others they are passed to where - this is annoyingly inconsistent
	}
}

func (a *Args) Values() ValueArgs {
	return a.values
}

func (a *Args) Wheres() WhereArgs {
	return a.wheres
}
