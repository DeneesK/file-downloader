package model

const (
	StatusCreated string = "created"
	StatusPending string = "pending"
	StatusRunning string = "running"
	StatusDone    string = "done"
	StatusFailed  string = "failed"
)

type Task struct {
	ID          string
	Status      string
	Archive     string
	Links       []string
	FailedLinks map[string]string
}
