package flow

type Status string

const (
	StatusNew       Status = "new"
	StatusRunning   Status = "running"
	StatusWaiting   Status = "waiting"
	StatusFailed    Status = "failed"
	StatusCompleted Status = "completed"
)

type Type string

const (
	TypeFunction Type = "function"
	TypeIO       Type = "io"
	TypeWait     Type = "wait"
)
