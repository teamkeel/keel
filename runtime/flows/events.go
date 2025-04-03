package flows

type FlowRunStarted struct {
	RunID string
}

type FlowRunUpdated struct {
	RunID        string `json:"runId"`
	RunCompleted bool   `json:"runCompleted"`
}
