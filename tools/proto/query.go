package proto

// FindByID finds a tool in the given tools message by id
func (tools *Tools) FindByID(id string) *ActionConfig {
	for _, t := range tools.Tools {
		if t.Id == id {
			return t
		}
	}

	return nil
}

// HasIDs checks that the tools wrapper contains all the tools with the given ids
func (tools *Tools) HasIDs(ids ...string) bool {
	for _, id := range ids {
		if exists := tools.FindByID(id); exists == nil {
			return false
		}
	}

	return true
}

// Diff will return a subset of the given tools which do not exist in our current tools wrapper
func (tools *Tools) Diff(others []*ActionConfig) []*ActionConfig {
	diffs := []*ActionConfig{}
	for _, t := range others {
		if tools.FindByID(t.Id) == nil {
			diffs = append(diffs, t)
		}
	}

	return diffs
}

// FindByAction will find in the given array the first tool config that has the required actionName
func FindByAction(tools []*ActionConfig, actionName string) *ActionConfig {
	for _, t := range tools {
		if t.ActionName == actionName {
			return t
		}
	}

	return nil
}

func (t *ActionConfig) HasInput(location *JsonPath) *RequestFieldConfig {
	for _, f := range t.GetInputs() {
		if f.GetFieldLocation().GetPath() == location.GetPath() {
			return f
		}
	}

	return nil
}

func (t *ActionConfig) HasResponse(location *JsonPath) *ResponseFieldConfig {
	for _, f := range t.GetResponse() {
		if f.GetFieldLocation().GetPath() == location.GetPath() {
			return f
		}
	}

	return nil
}

// DiffInputs will return a list of request field configs that exist in the given updated config but not in our receiver.
func (t *ActionConfig) DiffInputs(updated *ActionConfig) []*RequestFieldConfig {
	diff := []*RequestFieldConfig{}

	for _, i := range updated.GetInputs() {
		if exists := t.HasInput(i.GetFieldLocation()); exists == nil {
			diff = append(diff, i)
		}
	}

	return diff
}

// DiffResponse will return a list of response field configs that exist in the given updated config but not in our receiver.
func (t *ActionConfig) DiffResponse(updated *ActionConfig) []*ResponseFieldConfig {
	diff := []*ResponseFieldConfig{}

	for _, i := range updated.GetResponse() {
		if exists := t.HasResponse(i.GetFieldLocation()); exists == nil {
			diff = append(diff, i)
		}
	}

	return diff
}
