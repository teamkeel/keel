package actions

type ValueArgs map[string]any
type WhereArgs map[string]any

type Args struct {
	values ValueArgs
	wheres WhereArgs
}

func NewArgs(values ValueArgs, wheres WhereArgs) *Args {
	if values == nil {
		values = map[string]any{}
	}
	if wheres == nil {
		wheres = map[string]any{}
	}

	return &Args{
		// Used to provide data for the means of writing (create and update)
		values: values,
		// Used to filter data before performing an action (get, list, update, delete)
		wheres: wheres,
	}
}

func (a *Args) Values() ValueArgs {
	return a.values
}

func (a *Args) Wheres() WhereArgs {
	return a.wheres
}
