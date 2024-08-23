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
