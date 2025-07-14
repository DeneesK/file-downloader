package router

import (
	"context"

	"github.com/DeneesK/file-downloader/internal/app/model"
	"github.com/DeneesK/file-downloader/internal/app/router/middlewares"
	"github.com/go-chi/chi/v5"
)

type TaskService interface {
	CreateTask(ctx context.Context) (string, error)
	AddLinks(ctx context.Context, taskID string, links []string) error
	GetTask(ctx context.Context, taskID string) (*model.Task, error)
	GetNumberActiveTasks() int
}

type Logger interface {
	Infoln(args ...interface{})
	Errorf(template string, args ...interface{})
	Error(args ...interface{})
}

func NewRouter(taskService TaskService, log Logger) *chi.Mux {
	r := chi.NewRouter()

	loggingMiddleware := middlewares.NewLoggingMiddleware(log)
	r.Use(loggingMiddleware)
	r.Post("/task", CreateTask(taskService, log))
	r.Patch("/task/{id}", AddLinks(taskService, log))
	r.Get("/task/{id}", GetTask(taskService, log))
	return r
}
