package app

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/DeneesK/file-downloader/internal/app/model"
	"github.com/DeneesK/file-downloader/internal/app/router"
)

const shutdownTimeout = time.Second * 1

type Logger interface {
	Infoln(args ...interface{})
	Fatalf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Error(args ...interface{})
}

type TaskService interface {
	CreateTask(ctx context.Context) (string, error)
	AddLinks(ctx context.Context, taskID string, links []string) error
	GetTask(ctx context.Context, taskID string) (*model.Task, error)
	GetNumberActiveTasks() int
	Start()
	Shutdown()
}

type APP struct {
	srv         *http.Server
	log         Logger
	taskService TaskService
}

func NewApp(addr string, log Logger, taskService TaskService) *APP {
	r := router.NewRouter(taskService, log)
	s := http.Server{
		Addr:    addr,
		Handler: r,
	}
	return &APP{
		srv:         &s,
		log:         log,
		taskService: taskService,
	}
}

func (a *APP) Run() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGKILL,
	)
	defer stop()

	a.log.Infoln("starting application, server listening on", a.srv.Addr)

	go a.taskService.Start()

	go func() {
		err := a.srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			a.log.Fatalf("failed to start server: %s", err)
		}
	}()

	<-ctx.Done()

	a.log.Infoln("application shutdown process...")
	a.taskService.Shutdown()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := a.srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Error during shutdown: %s", err)
	}
	<-shutdownCtx.Done()
	a.log.Infoln("application and server gracefully stopped")
}
