package router

import (
	"net/http"
)

func CreateTask(taskService TaskService, log Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}
