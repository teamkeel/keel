package proto

func (a *Action) IsFunction() bool {
	return a.Implementation == ActionImplementation_ACTION_IMPLEMENTATION_CUSTOM
}

func (a *Action) IsArbitraryFunction() bool {
	return a.IsFunction() && (a.Type == ActionType_ACTION_TYPE_READ || a.Type == ActionType_ACTION_TYPE_WRITE)
}

func (a *Action) IsWriteAction() bool {
	switch a.Type {
	case ActionType_ACTION_TYPE_CREATE, ActionType_ACTION_TYPE_DELETE, ActionType_ACTION_TYPE_WRITE, ActionType_ACTION_TYPE_UPDATE:
		return true
	default:
		return false
	}
}

func (a *Action) IsReadAction() bool {
	switch a.Type {
	case ActionType_ACTION_TYPE_GET, ActionType_ACTION_TYPE_LIST, ActionType_ACTION_TYPE_READ:
		return true
	default:
		return false
	}
}

func (a *Action) IsUpdate() bool {
	return a.Type == ActionType_ACTION_TYPE_UPDATE
}

func (a *Action) IsCreate() bool {
	return a.Type == ActionType_ACTION_TYPE_CREATE
}

func (a *Action) IsList() bool {
	return a.Type == ActionType_ACTION_TYPE_LIST
}

func (a *Action) IsGet() bool {
	return a.Type == ActionType_ACTION_TYPE_GET
}

func (a *Action) IsDelete() bool {
	return a.Type == ActionType_ACTION_TYPE_DELETE
}

// FacetFields returns the fields that are used for faceting for this action
func FacetFields(schema *Schema, action *Action) []*Field {
	model := schema.FindModel(action.ModelName)
	var facetFields []*Field
	for _, name := range action.Facets {
		for _, field := range model.Fields {
			if field.Name == name {
				facetFields = append(facetFields, field)
				continue
			}
		}
	}
	return facetFields
}
