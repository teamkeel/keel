package flows

type FlowRunStarted struct {
	RunID string
}

type FlowRunUpdated struct {
	RunID        string
	RunCompleted bool
}
