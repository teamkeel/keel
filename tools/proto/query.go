package proto

// FindByID finds a tool in the given tools message by id
func (tools *Tools) FindByID(id string) *ActionConfig {
	if tools == nil {
		return nil
	}

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
func (tools *Tools) DiffIDs(ids []string) []string {
	diffs := []string{}
	for _, id := range ids {
		if tools.FindByID(id) == nil {
			diffs = append(diffs, id)
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

func (t *ActionConfig) FindInput(location *JsonPath) *RequestFieldConfig {
	for _, f := range t.GetInputs() {
		if f.GetFieldLocation().GetPath() == location.GetPath() {
			return f
		}
	}

	return nil
}

func (t *ActionConfig) FindInputByPath(location string) *RequestFieldConfig {
	for _, f := range t.GetInputs() {
		if f.GetFieldLocation().GetPath() == location {
			return f
		}
	}

	return nil
}

func (t *ActionConfig) FindResponse(location *JsonPath) *ResponseFieldConfig {
	for _, f := range t.GetResponse() {
		if f.GetFieldLocation().GetPath() == location.GetPath() {
			return f
		}
	}

	return nil
}

func (s *StringTemplate) Diff(other *StringTemplate) string {
	if other != nil {
		if s != nil && other.Template == s.Template {
			return ""
		}
		return other.Template
	}
	return ""
}

func (c *Capabilities) Diff(other *Capabilities) map[string]bool {
	diffs := map[string]bool{}
	if c.GetAudit() != other.GetAudit() {
		diffs["audit"] = other.GetAudit()
	}
	if c.GetComments() != other.GetComments() {
		diffs["comments"] = other.GetComments()
	}
	return diffs
}
