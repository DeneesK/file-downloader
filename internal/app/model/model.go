package model

const (
	StatusCreated string = "created"
	StatusRunning string = "running"
	StatusDone    string = "done"
	StatusFailed  string = "failed"
)

type Task struct {
	ID              string `json:"task_id"`
	Status          string `json:"status"`
	Archive         string `json:"archive,omitempty"`
	Links           []string
	DownloadedFiles []string
	FailedLinks     map[string]string `json:"failed_files,omitempty"`
}
