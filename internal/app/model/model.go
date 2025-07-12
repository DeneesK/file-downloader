package model

const (
	StatusPending string = "pending"
	StatusRunning string = "running"
	StatusDone    string = "done"
	StatusFailed  string = "failed"
)

type Task struct {
	ID      string
	Status  string
	Links   []string
	Archive string
}
