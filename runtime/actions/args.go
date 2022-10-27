package actions

// GraphQL, RPC, and the Testing framework all have subtle differences in how arguments are processed
// (JSON from RPC requests or from the Testing framework use ISO8601 for instance, whereas GraphQL has its own special datetime format)
// This interface enforces a contract that all of these API implementations must use to normalize the params
// The end result is that the actions code is not aware of these differences, and deals with native Go entities like time.Time (in the case of Dates / Timestamps)
// In the future, there will be other types of data to account for where the input structure differs dependent on the source

type ArgParser interface {
	Parse(input map[string]any) (values map[string]any, wheres map[string]any)
}

type Args struct {
	values map[string]any
	wheres map[string]any
}

func NewArgs(input map[string]any, parser ArgParser) *Args {
	values, wheres := parser.Parse(input)

	return &Args{
		values: values,
		wheres: wheres,
	}
}

func (a *Args) Values() map[string]any {

}

func (a *Args) Wheres() map[string]any {

}
